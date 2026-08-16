import urllib.request
import urllib.error
import json
import time

def trigger_trackers():
    print("Triggering GMP and Subscription tracking tasks for all active IPOs...")
    try:
        req = urllib.request.Request("http://localhost:8080/api/trackers/sync", method="POST")
        with urllib.request.urlopen(req) as response:
            if response.status == 200:
                resp_data = json.loads(response.read().decode())
                print(f"Success: {resp_data.get('message')}")
            else:
                print(f"Failed with status: {response.status}")
    except urllib.error.HTTPError as e:
        print(f"Failed: {e.code} - {e.reason}")
    except Exception as e:
        print(f"Exception: {e}")

if __name__ == "__main__":
    trigger_trackers()
