import urllib.request
import urllib.error
import json
import csv
import os

def audit_metrics():
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
        
    print(f"Auditing Metrics for {len(ipos)} IPOs...")
    
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
                    body = json.loads(response.read().decode())
                    metrics = body.get('metrics', {})
                    
                    if not metrics:
                        results.append({
                            "ID": ipo_id,
                            "Name": ipo_name,
                            "Status": "No Metrics Computed",
                            "RevCAGR3Y": 0.0, "PATCAGR3Y": 0.0, "EBITDAMargin": 0.0, 
                            "PATMargin": 0.0, "ROE": 0.0, "ROCE": 0.0, 
                            "DebtEquity": 0.0, "AssetTurnover": 0.0
                        })
                        continue
                    
                    results.append({
                        "ID": ipo_id,
                        "Name": ipo_name,
                        "Status": "Success",
                        "RevCAGR3Y": round(metrics.get('revenue_cagr_3y', 0.0), 4),
                        "PATCAGR3Y": round(metrics.get('pat_cagr_3y', 0.0), 4),
                        "EBITDAMargin": round(metrics.get('ebitda_margin', 0.0), 4),
                        "PATMargin": round(metrics.get('pat_margin', 0.0), 4),
                        "ROE": round(metrics.get('roe', 0.0), 4),
                        "ROCE": round(metrics.get('roce', 0.0), 4),
                        "DebtEquity": round(metrics.get('debt_to_equity', 0.0), 4),
                        "AssetTurnover": round(metrics.get('asset_turnover', 0.0), 4),
                    })
        except urllib.error.HTTPError as e:
            results.append({
                "ID": ipo_id,
                "Name": ipo_name,
                "Status": f"API Error: {e.code}",
                "RevCAGR3Y": 0.0, "PATCAGR3Y": 0.0, "EBITDAMargin": 0.0, 
                "PATMargin": 0.0, "ROE": 0.0, "ROCE": 0.0, 
                "DebtEquity": 0.0, "AssetTurnover": 0.0
            })
        except Exception as e:
            results.append({
                "ID": ipo_id,
                "Name": ipo_name,
                "Status": f"Exception: {str(e)}",
                "RevCAGR3Y": 0.0, "PATCAGR3Y": 0.0, "EBITDAMargin": 0.0, 
                "PATMargin": 0.0, "ROE": 0.0, "ROCE": 0.0, 
                "DebtEquity": 0.0, "AssetTurnover": 0.0
            })
            
    os.makedirs('reports', exist_ok=True)
    with open('reports/audit_metrics.csv', 'w', newline='') as f:
        fieldnames = ["ID", "Name", "Status", "RevCAGR3Y", "PATCAGR3Y", "EBITDAMargin", "PATMargin", "ROE", "ROCE", "DebtEquity", "AssetTurnover"]
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(results)
        
    print(f"\nMetrics audit complete! Check reports/audit_metrics.csv")

if __name__ == "__main__":
    audit_metrics()
