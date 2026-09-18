package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// WorkspaceMerger merges per-expert workspaces into a shared main workspace
// after each execution wave completes.
//
// WHY needed:
//   executeWaves() runs experts in parallel goroutines.
//   Each expert writes to its own isolated workspace to avoid git race conditions.
//   After the wave, their outputs must be merged into a single main workspace
//   so the next wave's experts can see all previous work.
//
// FLOW per wave:
//   1. Collect changed files from each expert workspace (git diff HEAD~1)
//   2. Detect conflicts (same file modified by multiple experts)
//   3. If conflicts: return error (caller decides: retry or fail wave)
//   4. If no conflicts: rsync each expert workspace into main/
//   5. Commit merged result in main/
//
// DIRECTORY STRUCTURE:
//   /workspaces/{workflow_id}/
//     main/                    <- merged result (seeded for next wave)
//     {expert_id_1}/           <- Expert 1 isolated workspace
//     {expert_id_2}/           <- Expert 2 isolated workspace
type WorkspaceMerger struct {
	logger *zap.Logger
}

// NewWorkspaceMerger creates a new WorkspaceMerger.
func NewWorkspaceMerger(logger *zap.Logger) *WorkspaceMerger {
	return &WorkspaceMerger{logger: logger}
}

// MergeWave merges all expert workspaces from a completed wave into main/.
//
// workflowWorkspace: /workspaces/{workflow_id}/
// expertIDs: IDs of experts whose workspaces need merging
//
// Returns error if merge conflicts detected.
func (m *WorkspaceMerger) MergeWave(
	ctx context.Context,
	workflowWorkspace string,
	expertIDs []string,
) error {
	if len(expertIDs) == 0 {
		return nil
	}

	// Single expert: no merge needed, just copy to main
	if len(expertIDs) == 1 {
		return m.copyToMain(ctx, workflowWorkspace, expertIDs[0])
	}

	// Multiple experts: detect conflicts first
	fileOwners := make(map[string][]string) // file -> [expert_ids that changed it]

	for _, expertID := range expertIDs {
		expertPath := filepath.Join(workflowWorkspace, expertID)
		files, err := m.getChangedFiles(ctx, expertPath)
		if err != nil {
			m.logger.Warn("workspace_merger: could not get changed files",
				zap.String("expert_id", expertID),
				zap.Error(err),
			)
			continue
		}
		for _, f := range files {
			fileOwners[f] = append(fileOwners[f], expertID)
		}
	}

	// Check for conflicts
	var conflicts []string
	for file, owners := range fileOwners {
		if len(owners) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("%s (modified by experts: %s)",
				file, strings.Join(owners, ", ")))
		}
	}
	if len(conflicts) > 0 {
		m.logger.Error("workspace_merger: merge conflicts detected",
			zap.Strings("conflicts", conflicts),
		)
		return fmt.Errorf("merge conflicts in wave: %s", strings.Join(conflicts, "; "))
	}

	// No conflicts: rsync each expert workspace into main/
	mainPath := filepath.Join(workflowWorkspace, "main")
	if err := os.MkdirAll(mainPath, 0755); err != nil {
		return fmt.Errorf("mkdir main workspace: %w", err)
	}

	for _, expertID := range expertIDs {
		expertPath := filepath.Join(workflowWorkspace, expertID)
		if err := m.rsyncToMain(ctx, expertPath, mainPath); err != nil {
			return fmt.Errorf("rsync expert %s to main: %w", expertID, err)
		}
		m.logger.Info("workspace_merger: merged expert workspace",
			zap.String("expert_id", expertID),
		)
	}

	// Commit merged result in main/
	if err := m.commitMerge(ctx, mainPath, len(expertIDs)); err != nil {
		return fmt.Errorf("commit merge: %w", err)
	}

	m.logger.Info("workspace_merger: wave merged successfully",
		zap.Int("experts", len(expertIDs)),
		zap.String("main", mainPath),
	)
	return nil
}

// getChangedFiles returns files changed in the last commit of a workspace.
func (m *WorkspaceMerger) getChangedFiles(ctx context.Context, workspacePath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "HEAD~1", "HEAD")
	cmd.Dir = workspacePath
	output, err := cmd.Output()
	if err != nil {
		// HEAD~1 may not exist if only one commit (initial)
		return nil, nil
	}
	var files []string
	for _, f := range strings.Split(string(output), "\n") {
		if f = strings.TrimSpace(f); f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// rsyncToMain copies files from expertPath into mainPath, excluding .git/
func (m *WorkspaceMerger) rsyncToMain(ctx context.Context, expertPath, mainPath string) error {
	cmd := exec.CommandContext(ctx, "rsync", "-a", "--exclude", ".git", expertPath+"/", mainPath+"/")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rsync: %w (output: %s)", err, string(output))
	}
	return nil
}

// copyToMain handles the single-expert case: just rsync to main.
func (m *WorkspaceMerger) copyToMain(ctx context.Context, workflowWorkspace, expertID string) error {
	expertPath := filepath.Join(workflowWorkspace, expertID)
	mainPath := filepath.Join(workflowWorkspace, "main")
	if err := os.MkdirAll(mainPath, 0755); err != nil {
		return fmt.Errorf("mkdir main workspace: %w", err)
	}
	return m.rsyncToMain(ctx, expertPath, mainPath)
}

// commitMerge stages and commits all changes in the main workspace.
func (m *WorkspaceMerger) commitMerge(ctx context.Context, mainPath string, expertCount int) error {
	// git add .
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = mainPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w (output: %s)", err, string(output))
	}

	// Check if there is anything to commit
	cmd = exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = mainPath
	if err := cmd.Run(); err == nil {
		// Nothing staged — no commit needed
		return nil
	}

	// git commit
	msg := fmt.Sprintf("feat(wave): merge %d expert workspaces", expertCount)
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", msg)
	cmd.Dir = mainPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w (output: %s)", err, string(output))
	}
	return nil
}
