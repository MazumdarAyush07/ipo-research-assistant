import sys
import urllib.request
import json
import time

def trigger_and_fetch_score(ipo_id):
    print(f"Triggering Scoring Engine for IPO ID {ipo_id}...")
    
    # 1. Trigger the score calculation
    trigger_url = f"http://localhost:8080/api/ipos/{ipo_id}/score/trigger"
    req = urllib.request.Request(trigger_url, method="POST")
    try:
        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read().decode('utf-8'))
            print("Score calculated successfully!")
            
            score_data = result.get('data', {})
            total = score_data.get('TotalScore', 0)
            rec = score_data.get('Recommendation', 'Unknown')
            
            print(f"\n======================================")
            print(f"        FINAL SCORE: {total}/100")
            print(f"     RECOMMENDATION: {rec}")
            print(f"======================================\n")
            
            print("BREAKDOWN:")
            print(f" - Financials (out of 40): {score_data.get('FinancialsScore', 0)} ({score_data.get('FinancialsReason', '')})")
            print(f" - Valuation  (out of 20): {score_data.get('ValuationScore', 0)} ({score_data.get('ValuationReason', '')})")
            print(f" - Promoter   (out of 10): {score_data.get('PromoterScore', 0)} ({score_data.get('PromoterReason', '')})")
            print(f" - Industry   (out of 10): {score_data.get('IndustryScore', 0)} ({score_data.get('IndustryReason', '')})")
            print(f" - Risk       (out of 10): {score_data.get('RiskScore', 0)} ({score_data.get('RiskReason', '')})")
            print(f" - Subs       (out of 5) : {score_data.get('SubscriptionScore', 0)} ({score_data.get('SubscriptionReason', '')})")
            print(f" - GMP        (out of 5) : {score_data.get('GmpScore', 0)} ({score_data.get('GmpReason', '')})")
            print("\n")

    except Exception as e:
        print(f"Failed to calculate score: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 trigger_score.py <ipo_id>")
        sys.exit(1)
    
    trigger_and_fetch_score(sys.argv[1])
