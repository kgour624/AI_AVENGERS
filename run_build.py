import subprocess, pathlib
out_path = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\build_output.txt")
print("starting build...")
p = subprocess.run(["docker","compose","build","migrate","--progress=plain"], capture_output=True, text=True, timeout=360, cwd=r"C:\Users\sharm\AI_AVENGERS-1")
out = (p.stdout or "") + "\n---STDERR---\n" + (p.stderr or "") + f"\nEXIT:{p.returncode}"
out_path.write_text(out, encoding="utf-8", errors="ignore")
print(out[-8000:])
print("written to", out_path, "exit", p.returncode)
