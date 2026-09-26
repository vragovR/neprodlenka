-- +goose Up
CREATE TABLE children (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    last_name   TEXT    NOT NULL,
    first_name  TEXT    NOT NULL,
    middle_name TEXT    NOT NULL DEFAULT '',
    price       INTEGER NOT NULL DEFAULT 13000 CHECK (price >= 0),
    active      INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1))
);

CREATE TABLE child_parents (
    user_id  INTEGER PRIMARY KEY,
    child_id INTEGER NOT NULL REFERENCES children (id)
);
CREATE INDEX child_parents_child_id ON child_parents (child_id);

CREATE TABLE no_school_days (
    date TEXT PRIMARY KEY
);

CREATE TABLE daily_posts (
    date       TEXT PRIMARY KEY,
    status     TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    message_id TEXT
);

CREATE TABLE attendance (
    date     TEXT    NOT NULL,
    child_id INTEGER NOT NULL REFERENCES children (id),
    status   TEXT    NOT NULL CHECK (status IN ('present', 'absent')),
    price    INTEGER NOT NULL CHECK (price >= 0),
    PRIMARY KEY (date, child_id)
);

CREATE TABLE report_months (
    month TEXT PRIMARY KEY,
    uid   TEXT NOT NULL UNIQUE
);

CREATE TABLE admins (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    login         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

-- +goose Down
DROP TABLE admins;
DROP TABLE report_months;
DROP TABLE attendance;
DROP TABLE daily_posts;
DROP TABLE no_school_days;
DROP TABLE child_parents;
DROP TABLE children;
