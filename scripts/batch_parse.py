import urllib.request
import urllib.error
import json
import time

def trigger_parsing_all():
    print("Fetching all IPOs...")
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
        
    print(f"Found {len(ipos)} IPOs. Triggering parsing jobs for all...")
    
    for ipo in ipos:
        ipo_id = ipo.get("ID")
        ipo_name = ipo.get("Name")
        if not ipo_id:
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
