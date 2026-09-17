CREATE TABLE fee_tables (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 year INTEGER NOT NULL CHECK(year BETWEEN 1 AND 9999),
 month INTEGER NOT NULL CHECK(month BETWEEN 1 AND 12),
 created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
 revision INTEGER NOT NULL DEFAULT 1 CHECK(revision>0)
);
CREATE TABLE fee_records (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 table_id INTEGER NOT NULL REFERENCES fee_tables(id) ON DELETE CASCADE,
 day INTEGER NOT NULL CHECK(day BETWEEN 1 AND 31),
 location1 TEXT NOT NULL, location2 TEXT NOT NULL,
 quantity_milli INTEGER NOT NULL, unit_price_cents INTEGER, amount_cents INTEGER NOT NULL, tag TEXT,
 CHECK(unit_price_cents IS NULL OR (unit_price_cents>=0 AND unit_price_cents%100=0))
);
CREATE INDEX records_order ON fee_records(table_id,day,id);
CREATE INDEX records_tag ON fee_records(table_id,tag,day,id);
CREATE TABLE locations (id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT NOT NULL UNIQUE,selection_count INTEGER NOT NULL DEFAULT 0);
CREATE TABLE tags (id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT NOT NULL UNIQUE,selection_count INTEGER NOT NULL DEFAULT 0);
PRAGMA application_id=1179931714;
PRAGMA user_version=1;
