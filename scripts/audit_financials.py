import urllib.request
import urllib.error
import json
import csv
import os

def check_value(val):
    if not val or not val.get('Valid'): return 0.0
    try:
        return float(val.get('String', '0.0'))
    except:
        return 0.0

def audit_all():
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
        
    print(f"Auditing {len(ipos)} IPOs...")
    
    results = []
    
    for ipo in ipos:
        ipo_id = ipo.get("ID")
        ipo_name = ipo.get("Name")
        if not ipo_id:
            continue
            
        print(f"Auditing IPO {ipo_id} - {ipo_name}...")
        try:
            req = urllib.request.Request(f"http://localhost:8080/api/ipos/{ipo_id}/financials")
            with urllib.request.urlopen(req) as response:
                if response.status == 200:
                    data = json.loads(response.read().decode()).get('data', [])
                    if not data:
                        results.append({
                            "ID": ipo_id,
                            "Name": ipo_name,
                            "Status": "No Financials Data",
                            "Rev": 0, "PAT": 0, "EBITDA": 0, "Assets": 0, "Debt": 0, "Equity": 0
                        })
                        continue
                    
                    latest = data[0]
                    results.append({
                        "ID": ipo_id,
                        "Name": ipo_name,
                        "Status": "Success",
                        "Rev": check_value(latest.get('Revenue')),
                        "PAT": check_value(latest.get('Pat')),
                        "EBITDA": check_value(latest.get('Ebitda')),
                        "Assets": check_value(latest.get('TotalAssets')),
                        "Debt": check_value(latest.get('TotalDebt')),
                        "Equity": check_value(latest.get('Equity')),
                    })
        except urllib.error.HTTPError as e:
            results.append({
                "ID": ipo_id,
                "Name": ipo_name,
                "Status": f"API Error: {e.code}",
                "Rev": 0, "PAT": 0, "EBITDA": 0, "Assets": 0, "Debt": 0, "Equity": 0
            })
        except Exception as e:
            results.append({
                "ID": ipo_id,
                "Name": ipo_name,
                "Status": f"Exception: {str(e)}",
                "Rev": 0, "PAT": 0, "EBITDA": 0, "Assets": 0, "Debt": 0, "Equity": 0
            })
            
    os.makedirs('reports', exist_ok=True)
    with open('reports/audit_financials.csv', 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=["ID", "Name", "Status", "Rev", "PAT", "EBITDA", "Assets", "Debt", "Equity"])
        writer.writeheader()
        writer.writerows(results)
        
    print(f"\nAudit complete! Check reports/audit_financials.csv")

if __name__ == "__main__":
    audit_all()
