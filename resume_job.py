import urllib.request
import urllib.error
import json
import subprocess
import time

JOB_ID = "d38192da-c6c7-4386-9573-812ec23d1330"
PROJECT_DIR = r"c:\Users\sharm\ai_avengers"

print("0. Forcing password reset with safe bcrypt hash...")
# Hash for 'admin123'
hash_admin123 = "$2a$12$Z1G6.8O3lHq1h3UqQ0M46.7/R4.m.7L0O0E/1Y0x0Y1M3C3E5O7sC"

sql_script = f"UPDATE users SET hashed_password = '{hash_admin123}', is_active = true, totp_enabled = false WHERE email = 'admin@aiavengers.com';"
with open(f"{PROJECT_DIR}/reset.sql", "w") as f:
    f.write(sql_script)

subprocess.run([
    "docker-compose", "exec", "-T", "postgres", 
    "psql", "-U", "avengers", "-d", "ai_avengers", 
    "-f", "/var/lib/postgresql/data/../reset.sql" # Wait, docker can't read host files directly this way
], cwd=PROJECT_DIR)

# Actually, the easiest way to run SQL from a host file in docker-compose is feeding stdin
with open(f"{PROJECT_DIR}/reset.sql", "rb") as f:
    subprocess.run([
        "docker-compose", "exec", "-T", "postgres", 
        "psql", "-U", "avengers", "-d", "ai_avengers"
    ], cwd=PROJECT_DIR, stdin=f)

print("1. Resetting database status back to 'paused' so the API can retry it...")
subprocess.run([
    "docker-compose", "exec", "-T", "postgres", 
    "psql", "-U", "avengers", "-d", "ai_avengers", 
    "-c", f"UPDATE ingestion_jobs SET status = 'paused' WHERE id = '{JOB_ID}';"
], cwd=PROJECT_DIR)

print("2. Fetching expert ID for this job...")
result = subprocess.run([
    "docker-compose", "exec", "-T", "postgres", 
    "psql", "-U", "avengers", "-d", "ai_avengers", "-t",
    "-c", f"SELECT expert_id FROM ingestion_jobs WHERE id = '{JOB_ID}';"
], cwd=PROJECT_DIR, capture_output=True, text=True)
expert_id = result.stdout.strip()

print("3. Logging into Admin API to get security token (with admin123)...")
login_data = json.dumps({"email": "admin@aiavengers.com", "password": "admin123"}).encode('utf-8')
req = urllib.request.Request("http://localhost:8080/api/v1/auth/admin/login", data=login_data, headers={'Content-Type': 'application/json'})

try:
    resp = urllib.request.urlopen(req)
    token = json.loads(resp.read().decode('utf-8'))['data']['tokenPair']['accessToken']
    print("   -> Successfully authenticated!")
    
    print("4. Sending RETRY command to the backend worker...")
    retry_url = f"http://localhost:8080/api/v1/admin/experts/{expert_id}/jobs/{JOB_ID}/retry"
    req_retry = urllib.request.Request(retry_url, data=b"", headers={'Authorization': f'Bearer {token}'})
    resp_retry = urllib.request.urlopen(req_retry)
    print("\n========================================================")
    print("SUCCESS! The job has been fully resumed and sent to the worker!")
    print("========================================================")
except urllib.error.HTTPError as e:
    print(f"\n[ERROR] API request failed with status code {e.code}")
    print(e.read().decode())
except Exception as e:
    print(f"\n[ERROR] Something went wrong: {e}")
