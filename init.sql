-- Initialize the events table and indexes
-- This script runs automatically when PostgreSQL starts

CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    order_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_timestamp BIGINT NOT NULL,
    received_at BIGINT NOT NULL,
    customer_id TEXT NOT NULL,
    restaurant_id TEXT NOT NULL,
    driver_id TEXT NOT NULL,
    location_lat DOUBLE PRECISION NOT NULL,
    location_lng DOUBLE PRECISION NOT NULL,
    platform_token TEXT NOT NULL,
    validation_status TEXT NOT NULL,
    validation_error TEXT
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_order_id ON events(order_id);
CREATE INDEX IF NOT EXISTS idx_event_timestamp ON events(event_timestamp);
CREATE INDEX IF NOT EXISTS idx_received_at ON events(received_at);
CREATE INDEX IF NOT EXISTS idx_validation_status ON events(validation_status);
CREATE INDEX IF NOT EXISTS idx_customer_id ON events(customer_id);
CREATE INDEX IF NOT EXISTS idx_restaurant_id ON events(restaurant_id);
