CREATE EXTENSION IF not EXISTS CITEXT;

CREATE UNLOGGED TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    nickname CITEXT COLLATE "C" NOT NULL UNIQUE,
    fullname TEXT NOT NULL,
    email CITEXT COLLATE "C" NOT NULL UNIQUE,
    about TEXT DEFAULT ''
);

CREATE INDEX users_email_idx on users using hash (email);
CREATE INDEX users_nickname_idx on users using hash (nickname);

CREATE UNLOGGED TABLE IF NOT EXISTS forums (
    id SERIAL NOT NULL PRIMARY KEY,
    slug CITEXT COLLATE "C" NOT NULL UNIQUE,
    title TEXT NOT NULL,
    user_nickname CITEXT COLLATE "C" NOT NULL,
    thread_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,

    FOREIGN KEY (user_nickname) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE UNIQUE INDEX forums_slug_idx on forums (slug);

CREATE UNLOGGED TABLE IF NOT EXISTS forum_user (
    forum_slug CITEXT COLLATE "C" NOT NULL,
    nickname CITEXT COLLATE "C" NOT NULL,
    UNIQUE (forum_slug, nickname),

    FOREIGN KEY (forum_slug) REFERENCES forums (slug) ON UPDATE CASCADE,
    FOREIGN KEY (nickname) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE INDEX forum_user_forum_nickname_idx on forum_user (forum_slug, nickname);
CREATE INDEX forum_user_nickname_idx on forum_user (nickname);

CREATE UNLOGGED TABLE IF NOT EXISTS threads (
    id SERIAL NOT NULL PRIMARY KEY,
    author CITEXT COLLATE "C" NOT NULL,
    created TIMESTAMPTZ,
    forum CITEXT COLLATE "C" NOT NULL,
    msg TEXT NOT NULL,
    title TEXT NOT NULL,
    slug CITEXT COLLATE "C" UNIQUE,
    votes INTEGER DEFAULT 0,

    FOREIGN KEY (author) REFERENCES users (nickname) ON UPDATE CASCADE,
    FOREIGN KEY (forum) REFERENCES forums (slug) ON UPDATE CASCADE
);

CREATE INDEX threads_slug_idx on threads using hash (slug);
CREATE INDEX threads_id_idx ON threads using hash (id);
CREATE INDEX threads_forum_idx ON threads using hash (forum);


CREATE UNLOGGED TABLE IF NOT EXISTS posts (
    id SERIAL NOT NULL PRIMARY KEY,
    author CITEXT COLLATE "C" NOT NULL,
    msg TEXT NOT NULL,
    parent INTEGER NOT NULL,
    thread INTEGER NOT NULL,
    thread_slug CITEXT COLLATE "C" NOT NULL,
    created TIMESTAMPTZ,
    forum CITEXT COLLATE "C" NOT NULL,
    isEdited BOOLEAN DEFAULT false,
    path INTEGER[] DEFAULT ARRAY []::INTEGER[],

    FOREIGN KEY (author) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE INDEX posts_path_idx ON posts (path);
CREATE INDEX posts_thread_idx ON posts (thread);
CREATE INDEX posts_id_idx ON posts (id);
CREATE INDEX posts_thread_id_path1_parent ON posts (thread, id, path[1], parent);

CREATE UNLOGGED TABLE IF NOT EXISTS thread_vote (
    thread_id INTEGER,
    nickname CITEXT COLLATE "C" NOT NULL,
    vote INTEGER NOT NULL,

    FOREIGN KEY (thread_id) REFERENCES threads (id),
    FOREIGN KEY (nickname) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE INDEX votes_user_thread ON thread_vote (thread_id, nickname);


CREATE OR REPLACE FUNCTION add_votes_to_thread() RETURNS TRIGGER AS
$add_votes_to_thread$
BEGIN
    UPDATE threads
    SET votes = votes + new.vote
    WHERE id = new.thread_id;

    RETURN new;
END;
$add_votes_to_thread$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_votes_in_thread() RETURNS TRIGGER AS
$update_votes_in_thread$
BEGIN
    UPDATE threads
    SET votes = votes - old.vote + new.vote
    WHERE id = new.thread_id;

    RETURN new;
END;
$update_votes_in_thread$ LANGUAGE plpgsql;

CREATE TRIGGER update_votes_on_insert
    AFTER INSERT
    ON thread_vote
    FOR EACH ROW
EXECUTE PROCEDURE add_votes_to_thread();

CREATE TRIGGER update_votes_on_update
    BEFORE UPDATE
    ON thread_vote
    FOR EACH ROW
EXECUTE PROCEDURE update_votes_in_thread();


CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
BEGIN
    IF (new.parent = 0) THEN
        new.path = new.path || new.id;
    ELSE
        new.path = (select path from posts where id = new.parent) || new.id;
    END IF;

    UPDATE forums SET post_count = post_count + 1 WHERE slug = new.forum;
    RETURN new;
end
$update_path$ LANGUAGE plpgsql;

CREATE TRIGGER update_post_path
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE update_path();


CREATE OR REPLACE FUNCTION update_thread_count() RETURNS TRIGGER AS
$update_thread_count$
BEGIN
    UPDATE forums SET thread_count = thread_count + 1 WHERE slug = new.forum;
    RETURN new;
end
$update_thread_count$ LANGUAGE plpgsql;

CREATE TRIGGER forum_thread_count
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE update_thread_count();


CREATE OR REPLACE FUNCTION add_user_to_forum() RETURNS TRIGGER AS
$add_user_to_forum$
BEGIN
    INSERT INTO forum_user (forum_slug, nickname)
    VALUES (new.forum, new.author)
    ON CONFLICT DO NOTHING;
    RETURN new;
END;
$add_user_to_forum$ LANGUAGE plpgsql;

CREATE TRIGGER add_user_to_forum_on_thread_creation
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE add_user_to_forum();

CREATE TRIGGER add_user_to_forum_on_post_creation
    AFTER INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE add_user_to_forum();
