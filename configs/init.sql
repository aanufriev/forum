CREATE UNLOGGED TABLE IF NOT EXISTS users (
    nickname TEXT NOT NULL UNIQUE PRIMARY KEY COLLATE "POSIX",
    fullname TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    about TEXT DEFAULT ''
);

CREATE UNIQUE INDEX email_unique_idx on users (LOWER(email));
CREATE UNIQUE INDEX nickname_unique_idx on users (LOWER(nickname));

CREATE UNLOGGED TABLE IF NOT EXISTS forums (
    slug TEXT NOT NULL UNIQUE PRIMARY KEY,
    title TEXT NOT NULL,
    user_nickname TEXT NOT NULL,
    thread_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,

    FOREIGN KEY (user_nickname) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE UNIQUE INDEX slug_unique_idx on forums (LOWER(slug));
CREATE INDEX forums_slug_idx on forums (slug, LOWER(slug));

CREATE UNLOGGED TABLE IF NOT EXISTS threads (
    id SERIAL NOT NULL PRIMARY KEY,
    author TEXT NOT NULL,
    created TIMESTAMPTZ,
    forum TEXT NOT NULL,
    msg TEXT NOT NULL,
    title TEXT NOT NULL,
    slug TEXT DEFAULT '',
    votes INTEGER DEFAULT 0,

    FOREIGN KEY (author) REFERENCES users (nickname) ON UPDATE CASCADE,
    FOREIGN KEY (forum) REFERENCES forums (slug) ON UPDATE CASCADE
);

CREATE UNIQUE INDEX thread_slug_unique_idx on threads (LOWER(slug));
CREATE UNIQUE INDEX thread_id_index ON threads (id);

CREATE UNLOGGED TABLE IF NOT EXISTS posts (
    id SERIAL NOT NULL PRIMARY KEY,
    author TEXT NOT NULL,
    msg TEXT NOT NULL,
    parent INTEGER NOT NULL,
    thread INTEGER NOT NULL,
    thread_slug TEXT NOT NULL,
    created TIMESTAMPTZ DEFAULT current_timestamp,
    forum TEXT NOT NULL,
    isEdited BOOLEAN DEFAULT false,

    FOREIGN KEY (author) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE UNLOGGED TABLE IF NOT EXISTS thread_vote (
    thread_id INTEGER,
    nickname TEXT NOT NULL,
    vote INTEGER NOT NULL,

    FOREIGN KEY (thread_id) REFERENCES threads (id),
    FOREIGN KEY (nickname) REFERENCES users (nickname) ON UPDATE CASCADE
);

CREATE INDEX votes_user_thread ON thread_vote (thread_id, LOWER(nickname));

CREATE FUNCTION add_votes_to_thread() RETURNS TRIGGER AS
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
