-- Deal Detection Engine Migration
-- Creates tables for route baselines, detected deals, and deal alerts

-- ============================================================================
-- 1. Route Baselines - Historical price statistics per route
-- ============================================================================
CREATE TABLE IF NOT EXISTS route_baselines (
    id SERIAL PRIMARY KEY,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    trip_length INTEGER NOT NULL,
    class VARCHAR(20) NOT NULL DEFAULT 'economy',
    
    -- Rolling statistics (updated periodically)
    sample_count INTEGER NOT NULL DEFAULT 0,
    mean_price DECIMAL(12,2),
    median_price DECIMAL(12,2),
    stddev_price DECIMAL(12,2),
    min_price DECIMAL(12,2),
    max_price DECIMAL(12,2),
    p10_price DECIMAL(12,2),  -- 10th percentile
    p25_price DECIMAL(12,2),  -- 25th percentile
    p75_price DECIMAL(12,2),  -- 75th percentile
    p90_price DECIMAL(12,2),  -- 90th percentile
    
    -- Time window for calculations
    window_start TIMESTAMP WITH TIME ZONE,
    window_end TIMESTAMP WITH TIME ZONE,
    
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(origin, destination, trip_length, class)
);

CREATE INDEX IF NOT EXISTS idx_route_baselines_route ON route_baselines(origin, destination);
CREATE INDEX IF NOT EXISTS idx_route_baselines_class ON route_baselines(class);
CREATE INDEX IF NOT EXISTS idx_route_baselines_updated ON route_baselines(updated_at);

-- ============================================================================
-- 2. Detected Deals - Raw deal detections with deduplication
-- ============================================================================
CREATE TABLE IF NOT EXISTS detected_deals (
    id SERIAL PRIMARY KEY,
    
    -- Route info
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    trip_length INTEGER,
    
    -- Price info
    price DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    baseline_mean DECIMAL(12,2),
    baseline_median DECIMAL(12,2),
    
    -- Deal scoring
    discount_percent DECIMAL(5,2),        -- How much below baseline (0-100)
    deal_score INTEGER,                   -- 0-100 composite score
    deal_classification VARCHAR(30),      -- 'good', 'great', 'amazing', 'error_fare'
    
    -- Cost-per-mile analysis
    distance_miles DECIMAL(10,2),
    cost_per_mile DECIMAL(10,4),
    cabin_class VARCHAR(20) DEFAULT 'economy',
    
    -- Source tracking
    source_type VARCHAR(20) NOT NULL DEFAULT 'sweep', -- 'sweep', 'social', 'webhook'
    source_id VARCHAR(64),                -- External ID from source
    search_url TEXT,
    
    -- Deduplication
    deal_fingerprint VARCHAR(64) NOT NULL, -- SHA256 of route+price_range+dates
    first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    times_seen INTEGER DEFAULT 1,
    
    -- Status
    status VARCHAR(20) DEFAULT 'active',  -- 'active', 'expired', 'published', 'verified'
    verified BOOLEAN DEFAULT FALSE,
    verified_price DECIMAL(12,2),
    verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(deal_fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_detected_deals_route ON detected_deals(origin, destination);
CREATE INDEX IF NOT EXISTS idx_detected_deals_status ON detected_deals(status);
CREATE INDEX IF NOT EXISTS idx_detected_deals_classification ON detected_deals(deal_classification);
CREATE INDEX IF NOT EXISTS idx_detected_deals_score ON detected_deals(deal_score DESC);
CREATE INDEX IF NOT EXISTS idx_detected_deals_discount ON detected_deals(discount_percent DESC);
CREATE INDEX IF NOT EXISTS idx_detected_deals_cpm ON detected_deals(cost_per_mile) WHERE cost_per_mile IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_detected_deals_expires ON detected_deals(expires_at) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_detected_deals_first_seen ON detected_deals(first_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_detected_deals_source ON detected_deals(source_type);

-- ============================================================================
-- 3. Deal Alerts - Published deals ready for notification
-- ============================================================================
CREATE TABLE IF NOT EXISTS deal_alerts (
    id SERIAL PRIMARY KEY,
    detected_deal_id INTEGER NOT NULL REFERENCES detected_deals(id) ON DELETE CASCADE,
    
    -- Snapshot of deal at publish time
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    discount_percent DECIMAL(5,2),
    deal_classification VARCHAR(30),
    deal_score INTEGER,
    
    -- Publishing
    published_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    publish_method VARCHAR(20) NOT NULL DEFAULT 'auto', -- 'auto', 'manual'
    
    -- Notification tracking (for future use)
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_sent_at TIMESTAMP WITH TIME ZONE,
    notification_channels TEXT[], -- 'email', 'webhook', 'push'
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deal_alerts_deal_id ON deal_alerts(detected_deal_id);
CREATE INDEX IF NOT EXISTS idx_deal_alerts_published ON deal_alerts(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_deal_alerts_not_sent ON deal_alerts(notification_sent) WHERE notification_sent = FALSE;

-- ============================================================================
-- 4. Deal Sources - For webhook integration with social-pulse
-- ============================================================================
CREATE TABLE IF NOT EXISTS deal_sources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    source_type VARCHAR(30) NOT NULL, -- 'social_pulse', 'external_api', 'manual'
    webhook_url TEXT,
    api_key VARCHAR(64),
    enabled BOOLEAN DEFAULT TRUE,
    last_received_at TIMESTAMP WITH TIME ZONE,
    deals_received INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default social-pulse source
INSERT INTO deal_sources (name, source_type, enabled)
VALUES ('social-pulse', 'social_pulse', true)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 5. Deal Config - Runtime configuration (allows hot updates)
-- ============================================================================
CREATE TABLE IF NOT EXISTS deal_config (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO deal_config (key, value, description) VALUES
    ('threshold_good', '0.20', 'Discount % for good deal classification'),
    ('threshold_great', '0.35', 'Discount % for great deal classification'),
    ('threshold_amazing', '0.50', 'Discount % for amazing deal classification'),
    ('threshold_error_fare', '0.70', 'Discount % for error fare classification'),
    ('cpm_economy', '0.05', 'Cost-per-mile threshold for economy class'),
    ('cpm_business', '0.10', 'Cost-per-mile threshold for business class'),
    ('cpm_first', '0.15', 'Cost-per-mile threshold for first class'),
    ('baseline_window_days', '0', 'Days of history for baseline calculation (0 = all history)'),
    ('baseline_min_samples', '10', 'Minimum samples required for baseline'),
    ('deal_ttl_hours', '24', 'Default deal expiration in hours'),
    ('auto_publish', 'true', 'Auto-publish deals meeting thresholds')
ON CONFLICT (key) DO NOTHING;
