-- Таблица posts
CREATE TABLE posts (
id SERIAL PRIMARY KEY,
user_id INTEGER NOT NULL,
title VARCHAR(255) NOT NULL,
content TEXT NOT NULL,
status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'published', 'failed')),
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- Таблица post_destinations
CREATE TABLE post_destinations (
id SERIAL PRIMARY KEY,
post_id INTEGER NOT NULL,
destination_name VARCHAR(20) NOT NULL CHECK (destination_name IN ('Telegram', 'VK', 'Dzen')),
scheduled_for TIMESTAMP WITH TIME ZONE,
published_at TIMESTAMP WITH TIME ZONE,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT fk_post_destinations_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE INDEX idx_post_destinations_post_id ON post_destinations(post_id);
CREATE INDEX idx_post_destinations_destination_name ON post_destinations(destination_name);
CREATE INDEX idx_post_destinations_scheduled_for ON post_destinations(scheduled_for);
CREATE INDEX idx_post_destinations_published_at ON post_destinations(published_at);
CREATE INDEX idx_post_destinations_created_at ON post_destinations(created_at);

