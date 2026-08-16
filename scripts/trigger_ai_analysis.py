import urllib.request
import urllib.error
import sys
import json

def trigger_analysis(ipo_id):
    url = f"http://localhost:8080/api/ipos/{ipo_id}/analysis/trigger"
    print(f"Triggering AI Analysis for IPO ID {ipo_id}...")
    
    req = urllib.request.Request(url, method='POST')
    try:
        with urllib.request.urlopen(req) as response:
            status = response.getcode()
            body = response.read().decode('utf-8')
            print("Success! Task enqueued.")
            print(json.loads(body))
            print("Check worker logs to see progress.")
    except urllib.error.HTTPError as e:
        print(f"Failed: {e.code}")
        print(e.read().decode('utf-8'))
    except Exception as e:
        print(f"Failed to connect: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 trigger_ai_analysis.py <ipo_id>")
        sys.exit(1)
    
    ipo_id = sys.argv[1]
    trigger_analysis(ipo_id)
