CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);
CREATE INDEX IF NOT EXISTS idx_users_username ON users USING gin (username gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_email ON users USING gin (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_followers ON followers USING btree (user_id, follower_id);
