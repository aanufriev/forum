CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE IF EXISTS forum_user CASCADE;
DROP TABLE IF EXISTS votes CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP FUNCTION IF EXISTS insert_thread_votes();
DROP FUNCTION IF EXISTS update_thread_votes();
DROP FUNCTION IF EXISTS set_post_path();
DROP FUNCTION IF EXISTS update_forum_threads();
DROP FUNCTION IF EXISTS update_forum_posts();
DROP FUNCTION IF EXISTS add_forum_user();

DROP TRIGGER IF EXISTS insert_thread_votes ON thread_vote;
DROP TRIGGER IF EXISTS update_thread_votes ON thread_vote;
DROP TRIGGER IF EXISTS set_post_path ON posts;
DROP TRIGGER IF EXISTS update_forum_threads ON threads;
DROP TRIGGER IF EXISTS update_forum_posts ON posts;
DROP TRIGGER IF EXISTS add_forum_user_new_thread ON threads;
DROP TRIGGER IF EXISTS add_forum_user_new_post ON posts;


CREATE UNLOGGED TABLE users(
    id SERIAL PRIMARY KEY,
    nickname CITEXT UNIQUE NOT NULL,
    fullname CITEXT NOT NULL,
    about TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL
);

CREATE INDEX index_users_nickname_hash ON users USING HASH (nickname);
CREATE INDEX index_users_email_hash ON users USING HASH (email);


CREATE UNLOGGED TABLE forums(
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    user_nickname CITEXT NOT NULL,
    post_count INT DEFAULT 0,
    thread_count INT DEFAULT 0,
    slug CITEXT UNIQUE NOT NULL,

    FOREIGN KEY (user_nickname) REFERENCES users (nickname)
);

CREATE INDEX index_forums ON forums (slug, title, user_nickname, post_count, thread_count);
CREATE INDEX index_forums_slug_hash ON forums USING hash (slug);
CREATE INDEX index_forums_users_foreign ON forums (user_nickname);


CREATE UNLOGGED TABLE threads(
    id SERIAL PRIMARY KEY,
    author CITEXT NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    forum CITEXT NOT NULL,
    msg TEXT NOT NULL,
    slug CITEXT UNIQUE,
    title CITEXT NOT NULL,
    votes INT DEFAULT 0,

    FOREIGN KEY (forum) REFERENCES forums (slug) ON DELETE CASCADE,
    FOREIGN KEY (author) REFERENCES users (nickname) ON DELETE CASCADE
);

CREATE INDEX index_threads_forum_created ON threads (forum, created);
CREATE INDEX index_threads_created ON threads (created);
CREATE INDEX index_threads_slug_hash ON threads USING HASH (slug);
CREATE INDEX index_threads_id_hash ON threads USING HASH (id);


CREATE UNLOGGED TABLE posts(
    id BIGSERIAL PRIMARY KEY,
    author CITEXT NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    forum CITEXT NOT NULL,
    isEdited BOOLEAN DEFAULT FALSE,
    msg TEXT NOT NULL,
    parent INT NOT NULL,
    thread INT NOT NULL,
    path BIGINT[],

    FOREIGN KEY (forum) REFERENCES forums (slug) ON DELETE CASCADE,
    FOREIGN KEY (author) REFERENCES users (nickname) ON DELETE CASCADE,
    FOREIGN KEY (thread) REFERENCES threads (id) ON DELETE CASCADE
);

CREATE INDEX index_posts_id on posts (id);
CREATE INDEX index_posts_thread_id on posts (thread, id);
CREATE INDEX index_posts_thread_path on posts (thread, path);
CREATE INDEX index_posts_thread_parent_path on posts (thread, parent, path);
CREATE INDEX index_posts_path1_path on posts ((path[1]), path);


CREATE UNLOGGED TABLE thread_vote(
    thread_id INT NOT NULL,
    vote INT NOT NULL,
    nickname CITEXT NOT NULL,

    FOREIGN KEY (thread_id) REFERENCES threads (id),
    FOREIGN KEY (nickname) REFERENCES users (nickname),
    UNIQUE (thread_id, nickname)
);

CREATE UNIQUE INDEX index_votes_user_thread ON thread_vote (thread_id, nickname);


CREATE UNLOGGED TABLE forum_user(
    forum_slug CITEXT NOT NULL,
    nickname CITEXT NOT NULL,
    FOREIGN KEY (forum_slug) REFERENCES forums (slug) ON DELETE CASCADE,
    FOREIGN KEY (nickname) REFERENCES users (nickname) ON DELETE CASCADE,
    UNIQUE (forum_slug, nickname)
);

CREATE INDEX index_forum_user ON forum_user (forum_slug, nickname);
CREATE INDEX index_forum_user_nickname ON forum_user (nickname);
cluster forum_user USING index_forum_user;


CREATE OR REPLACE FUNCTION insert_thread_votes()
    RETURNS TRIGGER AS
$insert_thread_votes$
BEGIN
    IF new.vote > 0 THEN
        UPDATE threads SET votes = (votes + 1)
        WHERE id = new.thread_id;
    ELSE
        UPDATE threads SET votes = (votes - 1)
        WHERE id = new.thread_id;
    END IF;
    RETURN new;
END;
$insert_thread_votes$ language plpgsql;

CREATE TRIGGER insert_thread_votes
    BEFORE INSERT
    ON thread_vote
    FOR EACH ROW
EXECUTE PROCEDURE insert_thread_votes();


CREATE OR REPLACE FUNCTION update_thread_votes()
    RETURNS TRIGGER AS
$update_thread_votes$
BEGIN
    IF new.vote > 0 THEN
        UPDATE threads
        SET votes = (votes + 2)
        WHERE threads.id = new.thread_id;
    else
        UPDATE threads
        SET votes = (votes - 2)
        WHERE threads.id = new.thread_id;
    END IF;
    RETURN new;
END;
$update_thread_votes$ LANGUAGE plpgsql;

CREATE TRIGGER update_thread_votes
    BEFORE UPDATE
    ON thread_vote
    FOR EACH ROW
EXECUTE PROCEDURE update_thread_votes();


CREATE OR REPLACE FUNCTION set_post_path()
    RETURNS TRIGGER AS
$set_post_path$
BEGIN
    IF (new.parent = 0) THEN
        new.path = new.path || new.id;
    ELSE
        new.path = (SELECT path FROM posts WHERE id = new.parent) || new.id;
    END IF;
    RETURN new;
END;
$set_post_path$ LANGUAGE plpgsql;

CREATE TRIGGER set_post_path
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE set_post_path();


CREATE OR REPLACE FUNCTION update_forum_threads()
    RETURNS TRIGGER AS
$update_forum_threads$
BEGIN
    UPDATE forums SET thread_count = thread_count + 1 WHERE slug = new.forum;
    RETURN new;
END;
$update_forum_threads$ LANGUAGE plpgsql;

CREATE TRIGGER update_forum_threads
    BEFORE INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE update_forum_threads();


CREATE OR REPLACE FUNCTION update_forum_posts()
    RETURNS TRIGGER AS
$update_forum_posts$
BEGIN
    UPDATE forums SET post_count = post_count + 1 WHERE slug = new.forum;
    RETURN new;
END;
$update_forum_posts$ LANGUAGE plpgsql;

CREATE TRIGGER update_forum_posts
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE update_forum_posts();


CREATE OR REPLACE FUNCTION add_forum_user()
    RETURNS TRIGGER AS
$add_forum_user$
BEGIN
    INSERT INTO forum_user (nickname, forum_slug)
    VALUES (new.author, new.forum)
    ON CONFLICT DO NOTHING;
    RETURN new;
END;
$add_forum_user$ LANGUAGE plpgsql;

CREATE TRIGGER add_forum_user_new_thread
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user();

CREATE TRIGGER add_forum_user_new_post
    AFTER INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user();
