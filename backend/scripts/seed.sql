-- Insert Sample IPO 1: TechNova Solutions (Mainboard)
INSERT INTO ipos (name, exchange_type, sector, price_band_low, price_band_high, open_date, close_date, listing_date, status, source_url)
VALUES ('TechNova Solutions Ltd', 'MAINBOARD', 'IT Services', 450.00, 475.00, '2026-08-10', '2026-08-12', '2026-08-18', 'UPCOMING', 'https://example.com/technova');

-- Insert Sample IPO 2: GreenFuture Energy (SME)
INSERT INTO ipos (name, exchange_type, sector, price_band_low, price_band_high, open_date, close_date, listing_date, status, source_url)
VALUES ('GreenFuture Energy SME', 'SME', 'Renewable Energy', 85.00, 85.00, '2026-07-01', '2026-07-03', '2026-07-10', 'LISTED', 'https://example.com/greenfuture');

-- Insert Financials for TechNova (IPO ID 1)
INSERT INTO financials (ipo_id, year, revenue, pat, ebitda, total_assets, total_debt, equity)
VALUES 
(1, 2024, 1200.50, 150.25, 200.00, 800.00, 100.00, 500.00),
(1, 2025, 1500.75, 200.50, 280.00, 1000.00, 80.00, 700.00),
(1, 2026, 1900.00, 300.00, 400.00, 1300.00, 50.00, 1000.00);

-- Insert Financials for GreenFuture (IPO ID 2)
INSERT INTO financials (ipo_id, year, revenue, pat, ebitda, total_assets, total_debt, equity)
VALUES 
(2, 2024, 25.50, 2.10, 4.00, 30.00, 15.00, 10.00),
(2, 2025, 45.00, 5.50, 8.50, 50.00, 20.00, 15.00),
(2, 2026, 80.20, 12.00, 18.00, 85.00, 18.00, 27.00);

-- Insert Valuation for TechNova
INSERT INTO valuation (ipo_id, issue_price, market_cap, pe_ratio, pb_ratio, ev_ebitda)
VALUES (1, 475.00, 5000.00, 16.66, 5.00, 12.50);

-- Insert Peer for TechNova
INSERT INTO peer_companies (ipo_id, name, ticker, pe, pb, ev_ebitda, roe, market_cap)
VALUES (1, 'TCS', 'TCS', 30.5, 12.0, 20.0, 40.0, 1400000.00);
