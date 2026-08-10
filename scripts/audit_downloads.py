import os
import glob
import pdfplumber

def audit_downloads():
    storage_dir = os.path.join(os.path.dirname(__file__), '..', 'storage')
    ipo_dirs = glob.glob(os.path.join(storage_dir, '*'))
    
    total_expected = len(ipo_dirs)
    downloaded = 0
    unexpected_eof = 0
    below_500kb = 0
    below_50_pages = 0
    
    print(f"Auditing {total_expected} IPO directories...")
    
    for ipo_dir in ipo_dirs:
        if not os.path.isdir(ipo_dir):
            continue
            
        pdf_path = os.path.join(ipo_dir, 'drhp.pdf')
        if not os.path.exists(pdf_path):
            print(f"Missing DRHP for {os.path.basename(ipo_dir)}")
            continue
            
        downloaded += 1
        size = os.path.getsize(pdf_path)
        if size < 500 * 1024:
            below_500kb += 1
            print(f"File below 500KB: {os.path.basename(ipo_dir)} ({size} bytes)")
            
        try:
            with pdfplumber.open(pdf_path) as pdf:
                page_count = len(pdf.pages)
                if page_count < 50:
                    below_50_pages += 1
                    print(f"File below 50 pages: {os.path.basename(ipo_dir)} ({page_count} pages)")
        except Exception as e:
            if "EOF" in str(e) or "Unexpected EOF" in str(e):
                unexpected_eof += 1
                print(f"Unexpected EOF: {os.path.basename(ipo_dir)}")
            else:
                print(f"Error opening {os.path.basename(ipo_dir)}: {e}")
                
    print("\n--- Audit Summary ---")
    print(f"Total Downloaded: {downloaded} / {total_expected}")
    print(f"Unexpected EOF: {unexpected_eof}")
    print(f"Below 500KB: {below_500kb}")
    print(f"Below 50 Pages: {below_50_pages}")

if __name__ == "__main__":
    audit_downloads()
