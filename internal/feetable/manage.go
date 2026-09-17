package feetable

import (
	"context"
	"database/sql"
)

// MergeTables moves records without recalculating money or changing record IDs.
// All versions and limits are checked before the first write.
func (s *Store) MergeTables(ctx context.Context, targetID int64, in MergeInput) (Table, error) {
	if targetID <= 0 || in.Revision <= 0 || len(in.Sources) < 1 || len(in.Sources) > 99 {
		return Table{}, invalid("每次请选择 2–100 张表格合并")
	}
	seen := map[int64]bool{targetID: true}
	for _, source := range in.Sources {
		if source.ID <= 0 || source.Revision <= 0 || seen[source.ID] {
			return Table{}, invalid("合并表格不能重复，且须包含有效版本")
		}
		seen[source.ID] = true
	}
	var out Table
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		target, err := readTable(ctx, tx, targetID)
		if err != nil {
			return err
		}
		if target.Revision != in.Revision {
			return conflict()
		}
		count := target.RecordCount
		for _, source := range in.Sources {
			t, err := readTable(ctx, tx, source.ID)
			if err != nil {
				return err
			}
			if t.Revision != source.Revision {
				return conflict()
			}
			if t.Year != target.Year || t.Month != target.Month {
				return invalid("只能合并年月相同的表格")
			}
			count += t.RecordCount
		}
		if count > MaxRecords {
			return invalid("合并后超过 5000 条记录，请减少所选表格")
		}
		for _, source := range in.Sources {
			if _, err = tx.ExecContext(ctx, "UPDATE fee_records SET table_id=? WHERE table_id=?", targetID, source.ID); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, "DELETE FROM fee_tables WHERE id=?", source.ID); err != nil {
				return err
			}
		}
		if err = touch(ctx, tx, targetID); err != nil {
			return err
		}
		out, err = readTable(ctx, tx, targetID)
		return err
	})
	return out, err
}

// Location edits affect the suggestion library; records keep their text snapshots.
func (s *Store) UpdateLocation(ctx context.Context, id int64, in LocationInput) error {
	name, err := cleanName(in.Name)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if err := checkLocation(ctx, tx, id, in.PreviousName); err != nil {
			return err
		}
		var duplicate bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM locations WHERE name=? AND id<>?)", name, id).Scan(&duplicate); err != nil {
			return err
		}
		if duplicate {
			return invalid("该地点已存在")
		}
		_, err := tx.ExecContext(ctx, "UPDATE locations SET name=? WHERE id=?", name, id)
		return err
	})
}

func (s *Store) DeleteLocation(ctx context.Context, id int64, previousName string) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if err := checkLocation(ctx, tx, id, previousName); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "DELETE FROM locations WHERE id=?", id)
		return err
	})
}

func checkLocation(ctx context.Context, tx *sql.Tx, id int64, previousName string) error {
	var name string
	err := tx.QueryRowContext(ctx, "SELECT name FROM locations WHERE id=?", id).Scan(&name)
	if err == sql.ErrNoRows {
		return missing()
	}
	if err != nil {
		return err
	}
	if name != previousName {
		return &Error{Code: "CONFLICT", Message: "地点已被修改，请刷新后核对"}
	}
	return nil
}
