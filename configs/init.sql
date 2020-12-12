CREATE TABLE IF NOT EXISTS users (
    nickname TEXT NOT NULL UNIQUE PRIMARY KEY,
    fullname TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    about TEXT DEFAULT ''
);

CREATE UNIQUE INDEX email_unique_idx on users (LOWER(email));
CREATE UNIQUE INDEX nickname_unique_idx on users (LOWER(nicknamel));

CREATE TABLE IF NOT EXISTS forums (
    slug TEXT NOT NULL UNIQUE PRIMARY KEY,
    title TEXT NOT NULL,
    user_nickname TEXT NOT NULL,

    FOREIGN KEY (user_nickname) REFERENCES users (nickname)
);

CREATE UNIQUE INDEX slug_unique_idx on forums (LOWER(slug));

CREATE TABLE IF NOT EXISTS threads (
    id SERIAL NOT NULL PRIMARY KEY,
    author TEXT NOT NULL,
    created TIMESTAMPTZ,
    forum TEXT NOT NULL,
    msg TEXT NOT NULL,
    title TEXT NOT NULL,
    slug TEXT,
    votes INTEGER DEFAULT 0,

    FOREIGN KEY (author) REFERENCES users (nickname),
    FOREIGN KEY (forum) REFERENCES forums (slug)
);

CREATE UNIQUE INDEX thread_slug_unique_idx on threads (LOWER(slug));

CREATE TABLE IF NOT EXISTS posts (
    id SERIAL NOT NULL PRIMARY KEY,
    author TEXT NOT NULL,
    msg TEXT NOT NULL,
    parent INTEGER NOT NULL,
    thread TEXT NOT NULL,
    created TIMESTAMPTZ,

    FOREIGN KEY (author) REFERENCES users (nickname)
);

CREATE TABLE IF NOT EXISTS thread_vote (
    thread_slug TEXT,
    thread_id INTEGER,
    nickname TEXT NOT NULL,
    vote INTEGER NOT NULL,

    FOREIGN KEY (nickname) REFERENCES users (nickname)
);
