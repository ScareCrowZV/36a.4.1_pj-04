CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    pub_time BIGINT NOT NULL,
    link TEXT UNIQUE NOT NULL
);

CREATE INDEX idx_posts_pub_time ON posts(pub_time DESC);