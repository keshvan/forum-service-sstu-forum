CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    user_id integer,
    username text NOT NULL,
    content text NOT NULL,
    created_at timestamp with time zone
)
