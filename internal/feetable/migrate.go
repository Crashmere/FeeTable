package feetable

import (
	"context"
	"database/sql"
)

// Migrate is an explicit offline operation: stop the server and back up first.
// All rows, IDs, revisions and sequence counters survive the schema change.
func Migrate(ctx context.Context, path string) error {
	s, err := Open(path, false)
	if err != nil {
		return err
	}
	defer s.Close()
	return s.transaction(ctx, func(tx *sql.Tx) error {
		var version int
		if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
			return err
		}
		if version == 2 {
			return nil
		}
		var sequence int64
		if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(seq),0) FROM sqlite_sequence WHERE name='fee_records'").Scan(&sequence); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
CREATE TABLE fee_records_v2 (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 table_id INTEGER NOT NULL REFERENCES fee_tables(id) ON DELETE CASCADE,
 day INTEGER NOT NULL CHECK(day BETWEEN 1 AND 31),
 location1 TEXT NOT NULL, location2 TEXT NOT NULL,
 quantity_milli INTEGER NOT NULL, unit_price_cents INTEGER, amount_cents INTEGER NOT NULL, tag TEXT,
 CHECK(unit_price_cents IS NULL OR unit_price_cents>=0)
);
INSERT INTO fee_records_v2 SELECT * FROM fee_records;
DROP TABLE fee_records;
ALTER TABLE fee_records_v2 RENAME TO fee_records;
CREATE INDEX records_order ON fee_records(table_id,day,id);
CREATE INDEX records_tag ON fee_records(table_id,tag,day,id);
PRAGMA user_version=2;
`)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE sqlite_sequence SET seq=MAX(seq,?) WHERE name='fee_records'", sequence); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO sqlite_sequence(name,seq) SELECT 'fee_records',? WHERE NOT EXISTS(SELECT 1 FROM sqlite_sequence WHERE name='fee_records')", sequence); err != nil {
			return err
		}
		return checkDB(ctx, tx)
	})
}
