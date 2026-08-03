-- +goose Up
-- +goose StatementBegin
CREATE TABLE ipos (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange_type VARCHAR(50),
    sector VARCHAR(100),
    price_band_low NUMERIC(10, 2),
    price_band_high NUMERIC(10, 2),
    open_date DATE,
    close_date DATE,
    listing_date DATE,
    status VARCHAR(50),
    source_url TEXT
);

CREATE TABLE financials (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    revenue NUMERIC(15, 2),
    pat NUMERIC(15, 2),
    ebitda NUMERIC(15, 2),
    total_assets NUMERIC(15, 2),
    total_debt NUMERIC(15, 2),
    equity NUMERIC(15, 2),
    UNIQUE(ipo_id, year)
);

CREATE TABLE valuation (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    issue_price NUMERIC(10, 2),
    market_cap NUMERIC(15, 2),
    pe_ratio NUMERIC(10, 2),
    pb_ratio NUMERIC(10, 2),
    ev_ebitda NUMERIC(10, 2),
    UNIQUE(ipo_id)
);

CREATE TABLE peer_companies (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    ticker VARCHAR(50),
    pe NUMERIC(10, 2),
    pb NUMERIC(10, 2),
    ev_ebitda NUMERIC(10, 2),
    roe NUMERIC(10, 2),
    market_cap NUMERIC(15, 2)
);

CREATE TABLE subscription_data (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    times_subscribed NUMERIC(10, 2),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gmp_history (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    gmp_amount NUMERIC(10, 2),
    premium_percent NUMERIC(10, 2),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    file_path TEXT NOT NULL,
    downloaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    parsed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE ai_analysis (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    business_model TEXT,
    moat TEXT,
    promoter_risk TEXT,
    legal_cases TEXT,
    customer_concentration TEXT,
    debt_assessment TEXT,
    key_risks TEXT,
    red_flags TEXT,
    management_assumptions TEXT,
    industry_outlook TEXT,
    raw_json JSONB,
    UNIQUE(ipo_id)
);

CREATE TABLE scores (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    financials_score NUMERIC(5, 2),
    financials_reason TEXT,
    valuation_score NUMERIC(5, 2),
    valuation_reason TEXT,
    promoter_score NUMERIC(5, 2),
    promoter_reason TEXT,
    industry_score NUMERIC(5, 2),
    industry_reason TEXT,
    risk_score NUMERIC(5, 2),
    risk_reason TEXT,
    subscription_score NUMERIC(5, 2),
    gmp_score NUMERIC(5, 2),
    final_score NUMERIC(5, 2),
    recommendation VARCHAR(50),
    scored_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ipo_id)
);

CREATE TABLE ipo_outcomes (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    issue_price NUMERIC(10, 2),
    listing_price NUMERIC(10, 2),
    listing_return_pct NUMERIC(10, 2),
    one_month_return_pct NUMERIC(10, 2),
    six_month_return_pct NUMERIC(10, 2),
    one_year_return_pct NUMERIC(10, 2),
    our_score_at_listing NUMERIC(5, 2),
    UNIQUE(ipo_id)
);

CREATE TABLE reports (
    id BIGSERIAL PRIMARY KEY,
    ipo_id BIGINT NOT NULL REFERENCES ipos(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    format VARCHAR(20) NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reports CASCADE;
DROP TABLE IF EXISTS ipo_outcomes CASCADE;
DROP TABLE IF EXISTS scores CASCADE;
DROP TABLE IF EXISTS ai_analysis CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS gmp_history CASCADE;
DROP TABLE IF EXISTS subscription_data CASCADE;
DROP TABLE IF EXISTS peer_companies CASCADE;
DROP TABLE IF EXISTS valuation CASCADE;
DROP TABLE IF EXISTS financials CASCADE;
DROP TABLE IF EXISTS ipos CASCADE;
-- +goose StatementEnd
