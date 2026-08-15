import urllib.request
import urllib.error
import json
import time
import subprocess
import os

def get_parsed_ipo_ids():
    db_url = os.environ.get("DATABASE_URL")
    if not db_url:
        print("DATABASE_URL not set. Make sure you run 'source .env' first.")
        return set()
    try:
        result = subprocess.run(
            ["psql", db_url, "-t", "-c", "SELECT ipo_id FROM ai_analysis;"],
            capture_output=True, text=True, check=True
        )
        return {int(x.strip()) for x in result.stdout.split('\n') if x.strip().isdigit()}
    except Exception as e:
        print(f"Failed to query database: {e}")
        return set()

def trigger_parsing_all():
    parsed_ids = get_parsed_ipo_ids()
    print(f"Found {len(parsed_ids)} IPOs already parsed in the database. Will skip them.")
    
    print("Fetching all IPOs from API...")
    try:
        req = urllib.request.Request("http://localhost:8080/api/ipos")
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            ipos = data.get("data", [])
            
        if not ipos:
            print("No IPOs found.")
            return
    except Exception as e:
        print(f"Failed to fetch IPOs: {e}")
        return
        
    print(f"Found {len(ipos)} total IPOs. Triggering parsing jobs for unparsed ones...")
    
    for ipo in ipos:
        ipo_id = ipo.get("ID")
        ipo_name = ipo.get("Name")
        if not ipo_id:
            continue
            
        if ipo_id in parsed_ids:
            print(f"Skipping IPO {ipo_id} - {ipo_name} (Already Parsed)")
            continue
            
        print(f"Triggering Parse for IPO {ipo_id} - {ipo_name}...")
        try:
            req = urllib.request.Request(f"http://localhost:8080/api/ipos/{ipo_id}/parse/trigger", method="POST")
            with urllib.request.urlopen(req) as response:
                if response.status == 200:
                    resp_data = json.loads(response.read().decode())
                    print(f"  Success: {resp_data.get('task_id')}")
        except urllib.error.HTTPError as e:
            if e.code == 404:
                print(f"  Skipped (No downloaded document found for IPO {ipo_id})")
            else:
                print(f"  Failed: {e.code} - {e.reason}")
        except Exception as e:
            print(f"  Exception: {e}")
            
        time.sleep(0.5)

if __name__ == "__main__":
    trigger_parsing_all()
