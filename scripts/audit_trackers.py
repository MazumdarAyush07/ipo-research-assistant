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
    print("Fetching all active IPOs...")
    try:
        req = urllib.request.Request("http://localhost:8080/api/ipos")
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            ipos = data.get("data", [])
        if not ipos:
            print("No IPOs found.")
            return
            
        # Filter for ACTIVE or UPCOMING
        active_ipos = [ipo for ipo in ipos if ipo.get('Status') and ipo['Status'].get('String') in ('ACTIVE', 'UPCOMING')]
    except Exception as e:
        print(f"Failed to fetch IPOs: {e}")
        return
        
    print(f"Auditing {len(active_ipos)} Active/Upcoming IPOs...")
    
    results = []
    
    for ipo in active_ipos:
        ipo_id = ipo.get("ID")
        ipo_name = ipo.get("Name")
        if not ipo_id:
            continue
            
        print(f"Auditing IPO {ipo_id} - {ipo_name}...")
        
        # 1. Fetch GMP
        gmp_status = "No GMP Data"
        gmp_amount = 0.0
        gmp_premium = 0.0
        try:
            req = urllib.request.Request(f"http://localhost:8080/api/ipos/{ipo_id}/gmp")
            with urllib.request.urlopen(req) as response:
                if response.status == 200:
                    data = json.loads(response.read().decode()).get('data', {})
                    if data and data.get('GmpAmount'):
                        gmp_status = "Success"
                        gmp_amount = check_value(data.get('GmpAmount'))
                        gmp_premium = check_value(data.get('PremiumPercent'))
        except urllib.error.HTTPError as e:
            if e.code != 404:
                gmp_status = f"API Error: {e.code}"
        except Exception as e:
            gmp_status = f"Exception: {str(e)}"
            
        # 2. Fetch Subscriptions
        sub_status = "No Sub Data"
        qib = 0.0
        nii = 0.0
        retail = 0.0
        total = 0.0
        
        try:
            req = urllib.request.Request(f"http://localhost:8080/api/ipos/{ipo_id}/subscriptions")
            with urllib.request.urlopen(req) as response:
                if response.status == 200:
                    data = json.loads(response.read().decode()).get('data', [])
                    if data:
                        sub_status = "Success"
                        for entry in data:
                            cat = entry.get('Category', '')
                            val = check_value(entry.get('TimesSubscribed'))
                            if cat == 'QIB': qib = val
                            elif cat == 'NII': nii = val
                            elif cat == 'Retail': retail = val
                            elif cat == 'Total': total = val
        except urllib.error.HTTPError as e:
            if e.code != 404:
                sub_status = f"API Error: {e.code}"
        except Exception as e:
            sub_status = f"Exception: {str(e)}"

        results.append({
            "ID": ipo_id,
            "Name": ipo_name,
            "GMP_Status": gmp_status,
            "GMP_Amount": gmp_amount,
            "GMP_Premium%": gmp_premium,
            "Sub_Status": sub_status,
            "QIB_x": qib,
            "NII_x": nii,
            "Retail_x": retail,
            "Total_x": total
        })
            
    os.makedirs('reports', exist_ok=True)
    with open('reports/audit_trackers.csv', 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=[
            "ID", "Name", "GMP_Status", "GMP_Amount", "GMP_Premium%", 
            "Sub_Status", "QIB_x", "NII_x", "Retail_x", "Total_x"
        ])
        writer.writeheader()
        writer.writerows(results)
        
    print(f"\nAudit complete! Check reports/audit_trackers.csv")

if __name__ == "__main__":
    audit_all()
