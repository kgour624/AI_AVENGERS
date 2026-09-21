package workflow

// The pure half of the client-git plumbing: deciding whether a remote may be
// touched at all, building the URL, and keeping the credential out of anything
// that gets logged. No database, no git process, no filesystem — stdlib only.
//
// Split out from client_repo.go for two reasons, in this order:
//
//  1. This is the security boundary for both §17 and §18. It is the code that
//     decides whether a client's write-scoped token is handed to a host, and it
//     is the code that decides whether that token appears in a log line. Code
//     with that job should be readable in one screen, on its own, without the
//     pgx and os/exec plumbing around it.
//  2. Because it depends on nothing but the standard library, it can be
//     exercised directly. The rest of this package cannot be built in an
//     environment with no module cache, so a file that CAN be is worth keeping
//     separate — the allowlist and the scrubber are exactly the two things that
//     must not be shipped on the strength of reading them.
//
// The rules themselves are stated once, in client_repo.go's file comment.

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrUnsupportedRemote means the requested git URL is not on a provider we
// support. A credential-safety check, not a convenience check: see rule 1 in
// client_repo.go.
var ErrUnsupportedRemote = errors.New("unsupported git remote")

// providerHosts is the allowlist. Only these two hosts may ever receive a
// client token from this codebase.
//
// Self-hosted GitLab is deliberately absent: internal/repo already hardcodes
// gitlab.com for its API calls (fetchGitLabTree, exchangeGitLabCode), so a
// self-hosted instance is not supported anywhere in the product today. Adding
// it here alone would let a token be pushed to a host the rest of the system
// cannot talk to — and would turn the allowlist into a list of one entry plus
// "anything the client types".
var providerHosts = map[string]string{
	"github": "github.com",
	"gitlab": "gitlab.com",
}

// gitCredentialUser is the username half of the basic-auth pair each provider
// expects when the password is a token. Both providers ignore the value but
// require it to be present.
var gitCredentialUser = map[string]string{
	"github": "x-access-token",
	"gitlab": "oauth2",
}

// remoteTarget is a validated client git remote.
type remoteTarget struct {
	Provider string // "github" | "gitlab"
	Host     string // from providerHosts, never from the input
	Path     string // "owner/repo", no leading slash, no .git suffix
}

// CleanURL is the remote with no credentials in it — safe to log, safe to store
// in an event, safe to show the client.
func (t remoteTarget) CleanURL() string {
	return fmt.Sprintf("https://%s/%s.git", t.Host, t.Path)
}

// authedURL is the remote with the token embedded. NEVER log this, never store
// it, never write it into a git config.
//
// Built through url.URL rather than fmt.Sprintf so the credential is escaped by
// the rules that actually apply to a URL's userinfo section. Sprintf with a raw
// token produces a broken URL the moment a provider issues a token containing
// ':' or '@' — and the failure would look like an authentication error, not
// like a quoting bug.
func (t remoteTarget) authedURL(token string) string {
	u := url.URL{
		Scheme: "https",
		User:   url.UserPassword(gitCredentialUser[t.Provider], token),
		Host:   t.Host,
		Path:   "/" + t.Path + ".git",
	}
	return u.String()
}

// parseRemoteTarget validates a client-supplied git URL against the provider
// allowlist and returns the pieces WE will use to rebuild the URL.
//
// The returned host comes from providerHosts, not from the input, so even a
// parse that somehow got past the comparison cannot redirect the push: we only
// ever concatenate a host we chose ourselves.
//
// Accepted inputs, all equivalent:
//
//	https://github.com/acme/thing
//	https://github.com/acme/thing.git
//	github.com/acme/thing
//	acme/thing                      (provider implied by the provider argument)
func parseRemoteTarget(provider, raw string) (remoteTarget, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	wantHost, ok := providerHosts[provider]
	if !ok {
		return remoteTarget{}, fmt.Errorf("%w: provider %q is not supported (github or gitlab)", ErrUnsupportedRemote, provider)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return remoteTarget{}, fmt.Errorf("%w: no repository given", ErrUnsupportedRemote)
	}

	// An ssh-style remote (git@host:owner/repo) cannot carry a token and this
	// container has no ssh key, so it is rejected with the real reason rather
	// than failing later inside git with a confusing message.
	if strings.HasPrefix(raw, "git@") || strings.HasPrefix(raw, "ssh://") {
		return remoteTarget{}, fmt.Errorf("%w: ssh remotes are not supported — use the https URL", ErrUnsupportedRemote)
	}

	// Does the input carry a host, or is it a bare owner/repo?
	//
	// The test is the FIRST segment only. Testing "does it contain a dot"
	// anywhere was tried first and was wrong: "acme/thing.git" contains a dot,
	// would have been treated as host-carrying, and a perfectly ordinary
	// owner/repo.git input was rejected with "acme is not github.com".
	//
	// Known edge: a GitLab group path may itself contain a dot ("my.org/proj"),
	// which this reads as a host and refuses. The refusal names the host it saw,
	// and passing the full https URL — what the UI sends — is unambiguous.
	firstSeg := raw
	if i := strings.Index(firstSeg, "/"); i >= 0 {
		firstSeg = firstSeg[:i]
	}
	path := raw
	if strings.Contains(raw, "://") || strings.Contains(firstSeg, ".") {
		// Looks like it carries a host. Normalise to something url.Parse
		// handles, then check the host.
		toParse := raw
		if !strings.Contains(toParse, "://") {
			toParse = "https://" + toParse
		}
		u, err := url.Parse(toParse)
		if err != nil {
			return remoteTarget{}, fmt.Errorf("%w: %q is not a URL", ErrUnsupportedRemote, raw)
		}
		if u.Scheme != "https" && u.Scheme != "http" {
			return remoteTarget{}, fmt.Errorf("%w: scheme %q is not allowed", ErrUnsupportedRemote, u.Scheme)
		}
		// u.Host carries any port and any userinfo is in u.User, so comparing
		// u.Host catches "github.com:8443" as well as a different host. A port
		// on a known host is still refused: the allowlist is the host exactly as
		// written in providerHosts.
		if !strings.EqualFold(u.Host, wantHost) {
			return remoteTarget{}, fmt.Errorf("%w: %s is not %s — a %s token is only ever sent to %s",
				ErrUnsupportedRemote, u.Host, wantHost, provider, wantHost)
		}
		// Any credentials the caller put in the URL are dropped here: we supply
		// the credential, the client does not get to choose it.
		path = u.Path
	}

	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	if path == "" {
		return remoteTarget{}, fmt.Errorf("%w: %q has no repository path", ErrUnsupportedRemote, raw)
	}
	segs := strings.Split(path, "/")
	if len(segs) < 2 {
		return remoteTarget{}, fmt.Errorf("%w: %q is not owner/repo", ErrUnsupportedRemote, raw)
	}
	for _, s := range segs {
		if s == "" || s == "." || s == ".." {
			return remoteTarget{}, fmt.Errorf("%w: %q has an empty or relative path segment", ErrUnsupportedRemote, raw)
		}
		// A segment carrying URL syntax means the input was crafted to be
		// re-parsed differently by git than it was by url.Parse.
		if strings.ContainsAny(s, "@:?#\\") {
			return remoteTarget{}, fmt.Errorf("%w: %q has an invalid character in the path", ErrUnsupportedRemote, raw)
		}
	}

	return remoteTarget{Provider: provider, Host: wantHost, Path: path}, nil
}

// validateBranchName rejects names git would reject, and names that would
// change the meaning of the refspec built from them.
//
// The push refspec is "refs/heads/<local>:refs/heads/<target>". A target
// containing a colon or a space does not produce an invalid refspec — it
// produces a DIFFERENT, valid one. That is the case worth checking for.
func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name is empty")
	}
	if strings.ContainsAny(name, " \t:?*[\\^~") {
		return fmt.Errorf("branch name %q contains a character git does not allow", name)
	}
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") ||
		strings.Contains(name, "..") || strings.Contains(name, "//") || strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name %q is not a valid git branch name", name)
	}
	return nil
}

// scrubSecrets removes credentials from text that is about to be logged or
// returned to a caller.
//
// Two passes, because either alone leaves a hole: the explicit secrets are
// removed by value (git echoes the push URL verbatim in some errors), and then
// any remaining `//user:pass@host` shape is flattened (git also rewrites the
// URL in its own messages, e.g. percent-encoding it, which makes the literal
// comparison miss).
func scrubSecrets(text string, secrets ...string) string {
	for _, s := range secrets {
		if len(s) < 4 {
			continue // too short to be a token; replacing it would mangle prose
		}
		text = strings.ReplaceAll(text, s, "***")
		text = strings.ReplaceAll(text, url.QueryEscape(s), "***")
	}
	return scrubURLCredentials(text)
}

// scrubURLCredentials rewrites every `scheme://anything@host` occurrence to
// `scheme://***@host`.
//
// Written as a scan rather than a regexp because the interesting case is
// pathological input (a token containing regexp metacharacters), and a scan has
// no backtracking behaviour to reason about.
func scrubURLCredentials(text string) string {
	const marker = "://"
	var b strings.Builder
	rest := text
	for {
		i := strings.Index(rest, marker)
		if i < 0 {
			b.WriteString(rest)
			return b.String()
		}
		afterScheme := i + len(marker)
		b.WriteString(rest[:afterScheme])
		rest = rest[afterScheme:]

		// The authority ends at the first / ? # or whitespace.
		end := len(rest)
		for j, r := range rest {
			if r == '/' || r == '?' || r == '#' || r == ' ' || r == '\n' || r == '\t' || r == '"' || r == '\'' {
				end = j
				break
			}
		}
		authority := rest[:end]
		if at := strings.LastIndex(authority, "@"); at >= 0 {
			b.WriteString("***@")
			b.WriteString(authority[at+1:])
		} else {
			b.WriteString(authority)
		}
		rest = rest[end:]
	}
}
