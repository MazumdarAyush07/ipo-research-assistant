-- Add subscription_reason and gmp_reason columns to the scores table.
-- These were computed by the scoring engine but were never persisted.

ALTER TABLE scores 
    ADD COLUMN IF NOT EXISTS subscription_reason TEXT,
    ADD COLUMN IF NOT EXISTS gmp_reason TEXT;
