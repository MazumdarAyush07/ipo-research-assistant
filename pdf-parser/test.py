from extractors.financials import parse_financials
import sys

def run():
    try:
        res = parse_financials("../storage/pramodini-medicare-ltd/drhp.pdf")
        print("RESULT:")
        print(res)
    except Exception as e:
        print("ERROR:", e)

if __name__ == "__main__":
    run()
