import subprocess
import time
import ctypes
import os
import threading

PROJECT_DIR = r"c:\Users\sharm\ai_avengers"

def _show_popup(message):
    ctypes.windll.user32.MessageBoxW(0, message, "Training Alert", 0x30 | 0x0)

def show_alert(message):
    print(f"\n[!!! ALERT !!!] {message}\n")
    # Windows popup alert (non-blocking)
    threading.Thread(target=_show_popup, args=(message,), daemon=True).start()

def get_job_status():
    try:
        # Run psql inside the postgres container so we don't need psycopg2!
        cmd = [
            "docker-compose", "exec", "-T", "postgres", 
            "psql", "-U", "avengers", "-d", "ai_avengers", 
            "-t", "-c", "SELECT id, status, error_message FROM ingestion_jobs ORDER BY created_at DESC LIMIT 1;"
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, cwd=PROJECT_DIR)
        
        output = result.stdout.strip()
        if output:
            parts = [p.strip() for p in output.split("|")]
            if len(parts) >= 2:
                job_id = parts[0]
                status = parts[1]
                error = parts[2] if len(parts) > 2 else ""
                return {"id": job_id, "status": status, "error": error}
    except Exception as e:
        return {"error": str(e)}
    return None

def check_logs():
    try:
        # Get last 150 lines of ml-sidecar and api logs
        result = subprocess.run(["docker-compose", "logs", "--tail=150", "api", "ml-sidecar"], 
                                capture_output=True, text=True, cwd=PROJECT_DIR)
        logs = result.stdout.lower()
        if "timeout" in logs or "network error" in logs or "connection refused" in logs:
            return "Network or Connection Error found in logs!"
        if "pause" in logs or "paused" in logs:
            return "Training pause detected in logs!"
    except Exception as e:
        pass
    return None

print("Starting live trace for Domain Expert training...")
print("Monitoring database and container logs every 30 seconds...\n")

while True:
    current_time = time.strftime('%H:%M:%S')
    print(f"--- Status Update ({current_time}) ---")
    
    # 1. Check DB Status
    db_status = get_job_status()
    if db_status:
        if "error" in db_status and len(db_status.keys()) == 1:
            print(f"Database Query Error: {db_status['error']}")
        else:
            print(f"Latest Job ID: {db_status['id']} | Status: {db_status['status']}")
            if db_status['status'] == 'failed' or db_status['error']:
                show_alert(f"Training Job Failed in DB!\nError: {db_status['error']}")
    else:
        print("No ingestion/training jobs found in database yet.")

    # 2. Check Logs for Network Errors
    log_alert = check_logs()
    if log_alert:
        show_alert(log_alert)

    print("Listening... (Next update in 30s)\n")
    time.sleep(30)
