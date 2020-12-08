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
