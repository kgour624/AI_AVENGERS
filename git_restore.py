import subprocess
import pathlib

file_path = "frontend/src/pages/admin/AdminRagSettings.tsx"
result = subprocess.run(["git", "show", f"HEAD:{file_path}"], capture_output=True, text=True, cwd=r"c:\Users\sharm\AI_AVENGERS-1")

if result.returncode == 0:
    pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1", file_path).write_text(result.stdout, encoding="utf-8")
    print("Successfully restored from git HEAD!")
else:
    print("Git error:", result.stderr)
