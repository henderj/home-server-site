PRAGMA foreign_keys=off;

BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS dice_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sides INTEGER NOT NULL,
    name VARCHAR(20),
    set_id INTEGER,
    FOREIGN KEY (set_id) REFERENCES dice_set(id) ON DELETE CASCADE
);

INSERT INTO dice_new (id, sides, name, set_id) SELECT id, sides, name, set_id FROM dice;

DROP TABLE dice;

ALTER TABLE dice_new RENAME TO dice;

CREATE TABLE IF NOT EXISTS roll_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    die_id INTEGER,
    value INTEGER,
    FOREIGN KEY (die_id) REFERENCES dice(id) ON DELETE CASCADE
);

INSERT INTO roll_new (id, die_id, value) SELECT id, die_id, value FROM roll;

DROP TABLE roll;

ALTER TABLE roll_new RENAME TO roll;

COMMIT;

PRAGMA foreign_keys=on;
