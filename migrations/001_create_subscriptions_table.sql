CREATE TABLE subscriptions (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               service_name VARCHAR(255) NOT NULL,
                               price INTEGER NOT NULL CHECK (price >= 0),
                               user_id UUID NOT NULL,
                               start_date VARCHAR(7) NOT NULL,
                               end_date VARCHAR(7),
                               created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_id ON subscriptions(user_id);
CREATE INDEX idx_service_name ON subscriptions(service_name);