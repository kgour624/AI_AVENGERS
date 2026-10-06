import subprocess, sys

try:
    res = subprocess.run(["go", "build", "./internal/receptionist"], cwd=r"C:\Users\sharm\AI_AVENGERS-1\backend-go", capture_output=True, text=True)
    print("STDOUT:", res.stdout)
    print("STDERR:", res.stderr)
    print("EXIT:", res.returncode)
except Exception as e:
    print("ERROR:", e)
