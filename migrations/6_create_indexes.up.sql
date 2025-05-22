CREATE INDEX IF NOT EXISTS idx_topics_category_id ON public.topics(category_id);

CREATE INDEX IF NOT EXISTS idx_topics_author_id ON public.topics(author_id);

CREATE INDEX IF NOT EXISTS idx_posts_topic_id ON public.posts(topic_id);

CREATE INDEX IF NOT EXISTS idx_posts_author_id ON public.posts(author_id);

CREATE INDEX IF NOT EXISTS idx_posts_reply_to ON public.posts(reply_to);