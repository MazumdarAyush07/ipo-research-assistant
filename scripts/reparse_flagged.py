#!/usr/bin/env python3
"""
Targeted re-parse script for 6 flagged IPOs with data quality issues.
Skips all healthy IPOs and re-queues only the ones that need fixing.
"""
import os
import urllib.request
import json


DATABASE_URL = os.environ.get("DATABASE_URL")
API_BASE = "http://localhost:8080"

# IPOs flagged in the audit for bad data quality
FLAGGED_IDS = {
    26,  # Dhanwel Hybrid Seeds IPO       - Assets = 0.0
    28,  # Fascinate Textiles IPO          - Assets = 0.0
    29,  # Technocrats Plasma Systems IPO  - Revenue 100x too large (Lakhs not Crores)
    30,  # ENS Enterprises IPO             - Assets = 0.0, suspiciously large revenue
    46,  # Propshop Events & Exhibitions   - Assets = 0.0, Equity = 0.0
    53,  # Millworks Technologies IPO      - Assets = 0.0
}

def main():
    if not DATABASE_URL:
        print("ERROR: DATABASE_URL not set. Run: set -a; source .env; set +a")
        return

    print(f"Re-parsing {len(FLAGGED_IDS)} flagged IPOs with improved extraction...")
    print(f"Flagged IDs: {sorted(FLAGGED_IDS)}\n")

    # Fetch all IPOs from the API
    req = urllib.request.Request(f"{API_BASE}/api/ipos")
    with urllib.request.urlopen(req) as response:
        ipos_resp = json.loads(response.read().decode())
    ipos = ipos_resp.get("data", []) if isinstance(ipos_resp, dict) else ipos_resp

    triggered = 0
    for ipo in ipos:
        ipo_id = ipo.get("ID")
        name = ipo.get("Name", f"IPO #{ipo_id}")

        if ipo_id not in FLAGGED_IDS:
            continue

        # Force re-parse by hitting the parse endpoint directly
        req = urllib.request.Request(f"{API_BASE}/api/ipos/{ipo_id}/parse/trigger", method="POST")
        try:
            with urllib.request.urlopen(req) as trigger_resp:
                if trigger_resp.status == 200:
                    data = json.loads(trigger_resp.read().decode())
                    task_id = data.get("task_id", "unknown")
                    print(f"  ✅ Triggered re-parse for IPO {ipo_id} - {name}")
                    print(f"     Task ID: {task_id}")
                    triggered += 1
        except Exception as e:
            print(f"  ❌ Failed to trigger IPO {ipo_id} - {name}: {e}")

    print(f"\nDone! Triggered {triggered}/{len(FLAGGED_IDS)} re-parse jobs.")
    print("Watch docker-compose logs to track progress.")
    print("Once complete, re-run: python3 scripts/audit_financials.py")

if __name__ == "__main__":
    main()
