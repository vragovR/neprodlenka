-- +goose Up
CREATE TABLE child_parents_new (
    user_id  INTEGER NOT NULL,
    child_id INTEGER NOT NULL REFERENCES children (id),
    PRIMARY KEY (user_id, child_id)
);
INSERT INTO child_parents_new (user_id, child_id) SELECT user_id, child_id FROM child_parents;
DROP TABLE child_parents;
ALTER TABLE child_parents_new RENAME TO child_parents;
CREATE INDEX child_parents_child_id ON child_parents (child_id);

-- +goose Down
DELETE FROM child_parents WHERE rowid NOT IN (SELECT MIN(rowid) FROM child_parents GROUP BY user_id);
CREATE TABLE child_parents_old (
    user_id  INTEGER PRIMARY KEY,
    child_id INTEGER NOT NULL REFERENCES children (id)
);
INSERT INTO child_parents_old (user_id, child_id) SELECT user_id, child_id FROM child_parents;
DROP TABLE child_parents;
ALTER TABLE child_parents_old RENAME TO child_parents;
CREATE INDEX child_parents_child_id ON child_parents (child_id);
