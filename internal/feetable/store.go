package feetable

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

const tableSelect = `SELECT t.id,t.year,t.month,t.created_at,t.updated_at,t.revision,COUNT(r.id),COALESCE(SUM(r.amount_cents),0) FROM fee_tables t LEFT JOIN fee_records r ON r.table_id=t.id`

type scanner interface{ Scan(...any) error }

func scanTable(row scanner) (Table, error) {
	var t Table
	var total int64
	err := row.Scan(&t.ID, &t.Year, &t.Month, &t.CreatedAt, &t.UpdatedAt, &t.Revision, &t.RecordCount, &total)
	t.Total = FormatDecimal(total, 2)
	if err == sql.ErrNoRows {
		err = missing()
	}
	return t, err
}
func readTable(ctx context.Context, q queryer, id int64) (Table, error) {
	return scanTable(q.QueryRowContext(ctx, tableSelect+" WHERE t.id=? GROUP BY t.id", id))
}
func (s *Store) Tables(ctx context.Context, in TableQuery) (TableList, error) {
	if in.Page < 1 || in.Page > 1000000 {
		return TableList{}, invalid("页码无效")
	}
	order := "t.updated_at DESC,t.id DESC"
	switch in.Sort {
	case "", "updated_desc":
	case "updated_asc":
		order = "t.updated_at,t.id"
	case "month_desc":
		order = "t.year DESC,t.month DESC,t.updated_at DESC,t.id DESC"
	case "month_asc":
		order = "t.year,t.month,t.updated_at DESC,t.id DESC"
	default:
		return TableList{}, invalid("排序方式无效")
	}
	where := ""
	args := []any{}
	if in.Year != 0 || in.Month != 0 {
		if e := validMonth(in.Year, in.Month); e != nil {
			return TableList{}, e
		}
		where = " WHERE t.year=? AND t.month=?"
		args = append(args, in.Year, in.Month)
	}
	out := TableList{Items: []Table{}, Page: in.Page, PageSize: 30}
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		if e := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM fee_tables t"+where, args...).Scan(&out.Total); e != nil {
			return e
		}
		rows, e := tx.QueryContext(ctx, tableSelect+where+" GROUP BY t.id ORDER BY "+order+" LIMIT 30 OFFSET ?", append(args, (in.Page-1)*30)...)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			t, e := scanTable(rows)
			if e != nil {
				return e
			}
			out.Items = append(out.Items, t)
		}
		return rows.Err()
	})
	return out, err
}
func (s *Store) CreateTable(ctx context.Context, in TableInput) (Table, error) {
	if err := validMonth(in.Year, in.Month); err != nil {
		return Table{}, err
	}
	var out Table
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		now := time.Now().UnixMilli()
		res, e := tx.ExecContext(ctx, "INSERT INTO fee_tables(year,month,created_at,updated_at) VALUES(?,?,?,?)", in.Year, in.Month, now, now)
		if e != nil {
			return e
		}
		id, e := res.LastInsertId()
		if e != nil {
			return e
		}
		out, e = readTable(ctx, tx, id)
		return e
	})
	return out, err
}
func (s *Store) UpdateTable(ctx context.Context, id int64, in TableInput) (Table, error) {
	if err := validMonth(in.Year, in.Month); err != nil {
		return Table{}, err
	}
	var out Table
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		t, e := readTable(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.Revision != in.Revision {
			return conflict()
		}
		var maxDay int
		if e = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(day),1) FROM fee_records WHERE table_id=?", id).Scan(&maxDay); e != nil {
			return e
		}
		if e = validDay(in.Year, in.Month, maxDay); e != nil {
			return invalid("已有记录的日期不适用于新月份，请先调整记录")
		}
		_, e = tx.ExecContext(ctx, "UPDATE fee_tables SET year=?,month=?,updated_at=?,revision=revision+1 WHERE id=?", in.Year, in.Month, time.Now().UnixMilli(), id)
		if e != nil {
			return e
		}
		out, e = readTable(ctx, tx, id)
		return e
	})
	return out, err
}
func (s *Store) DeleteTable(ctx context.Context, id, revision int64) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		t, e := readTable(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.Revision != revision {
			return conflict()
		}
		_, e = tx.ExecContext(ctx, "DELETE FROM fee_tables WHERE id=?", id)
		return e
	})
}
func scanRecord(row scanner) (Record, error) {
	var r Record
	var quantity, amount int64
	var price sql.NullInt64
	var tag sql.NullString
	e := row.Scan(&r.ID, &r.TableID, &r.Day, &r.Location1, &r.Location2, &quantity, &price, &amount, &tag)
	r.Quantity = FormatDecimal(quantity, 3)
	r.Amount = FormatDecimal(amount, 2)
	if price.Valid {
		v := FormatDecimal(price.Int64, 2)
		r.UnitPrice = &v
	}
	if tag.Valid {
		r.Tag = &tag.String
	}
	if e == sql.ErrNoRows {
		e = missing()
	}
	return r, e
}

const recordSelect = "SELECT id,table_id,day,location1,location2,quantity_milli,unit_price_cents,amount_cents,tag FROM fee_records"

func (s *Store) Report(ctx context.Context, id int64, tag *string, page int, all bool) (Report, error) {
	out := Report{Records: []Record{}, Tags: []string{}, FilterTag: tag, Page: page, PageSize: 50}
	if page < 1 {
		return out, invalid("页码无效")
	}
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		var e error
		out.Table, e = readTable(ctx, tx, id)
		if e != nil {
			return e
		}
		rows, e := tx.QueryContext(ctx, "SELECT DISTINCT tag FROM fee_records WHERE table_id=? AND tag IS NOT NULL ORDER BY tag", id)
		if e != nil {
			return e
		}
		for rows.Next() {
			var value string
			if e = rows.Scan(&value); e != nil {
				rows.Close()
				return e
			}
			out.Tags = append(out.Tags, value)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		where := " WHERE table_id=?"
		args := []any{id}
		if tag != nil {
			where += " AND tag=?"
			args = append(args, *tag)
		}
		var total int64
		if e = tx.QueryRowContext(ctx, "SELECT COUNT(*),COALESCE(SUM(amount_cents),0) FROM fee_records"+where, args...).Scan(&out.Count, &total); e != nil {
			return e
		}
		out.Total = FormatDecimal(total, 2)
		if out.Count > MaxRecords {
			return invalid("表格记录数超出限制")
		}
		query := recordSelect + where + " ORDER BY day,id"
		if !all {
			query += " LIMIT 50 OFFSET ?"
			args = append(args, (page-1)*50)
		} else {
			out.PageSize = out.Count
		}
		rows, e = tx.QueryContext(ctx, query, args...)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			r, e := scanRecord(rows)
			if e != nil {
				return e
			}
			out.Records = append(out.Records, r)
		}
		return rows.Err()
	})
	return out, err
}
func (s *Store) SaveRecord(ctx context.Context, tableID, id int64, in RecordInput) (Record, error) {
	var out Record
	var err error
	in.Location1, err = cleanName(in.Location1)
	if err != nil {
		return out, err
	}
	in.Location2, err = cleanName(in.Location2)
	if err != nil {
		return out, err
	}
	if in.Tag != nil {
		if strings.TrimSpace(*in.Tag) == "" {
			in.Tag = nil
		} else {
			v, e := cleanName(*in.Tag)
			if e != nil {
				return out, e
			}
			in.Tag = &v
		}
	}
	quantity, err := ParseDecimal(in.Quantity, 3)
	if err != nil {
		return out, err
	}
	var price *int64
	var amount int64
	if in.UnitPrice != nil {
		p, e := ParseDecimal(*in.UnitPrice, 2)
		if e != nil {
			return out, invalid("单价：" + e.Error())
		}
		if p < 0 {
			return out, invalid("单价不能为负数")
		}
		price = &p
		amount, err = Calculate(quantity, p)
	} else {
		amount, err = ParseDecimal(in.Amount, 2)
	}
	if err != nil {
		return out, err
	}
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		t, e := readTable(ctx, tx, tableID)
		if e != nil {
			return e
		}
		if t.Revision != in.Revision {
			return conflict()
		}
		if e = validDay(t.Year, t.Month, in.Day); e != nil {
			return e
		}
		increment := 0
		if id == 0 {
			if t.RecordCount >= MaxRecords {
				return invalid("每张表最多 5000 条记录，请新建表格")
			}
			res, e := tx.ExecContext(ctx, "INSERT INTO fee_records(table_id,day,location1,location2,quantity_milli,unit_price_cents,amount_cents,tag) VALUES(?,?,?,?,?,?,?,?)", tableID, in.Day, in.Location1, in.Location2, quantity, price, amount, in.Tag)
			if e != nil {
				return e
			}
			id, e = res.LastInsertId()
			if e != nil {
				return e
			}
			increment = 1
		} else {
			res, e := tx.ExecContext(ctx, "UPDATE fee_records SET day=?,location1=?,location2=?,quantity_milli=?,unit_price_cents=?,amount_cents=?,tag=? WHERE id=? AND table_id=?", in.Day, in.Location1, in.Location2, quantity, price, amount, in.Tag, id, tableID)
			if e != nil {
				return e
			}
			n, e := res.RowsAffected()
			if e != nil {
				return e
			}
			if n == 0 {
				return missing()
			}
		}
		for _, name := range []string{in.Location1, in.Location2} {
			if _, e = tx.ExecContext(ctx, "INSERT INTO locations(name,selection_count) VALUES(?,?) ON CONFLICT(name) DO UPDATE SET selection_count=selection_count+excluded.selection_count", name, increment); e != nil {
				return e
			}
		}
		if in.Tag != nil {
			if _, e = tx.ExecContext(ctx, "INSERT INTO tags(name,selection_count) VALUES(?,?) ON CONFLICT(name) DO UPDATE SET selection_count=selection_count+excluded.selection_count", *in.Tag, increment); e != nil {
				return e
			}
		}
		if e = touch(ctx, tx, tableID); e != nil {
			return e
		}
		out, e = scanRecord(tx.QueryRowContext(ctx, recordSelect+" WHERE id=?", id))
		return e
	})
	return out, err
}
func touch(ctx context.Context, tx *sql.Tx, id int64) error {
	_, e := tx.ExecContext(ctx, "UPDATE fee_tables SET updated_at=?,revision=revision+1 WHERE id=?", time.Now().UnixMilli(), id)
	return e
}
func (s *Store) DeleteRecord(ctx context.Context, tableID, id, revision int64) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		t, e := readTable(ctx, tx, tableID)
		if e != nil {
			return e
		}
		if t.Revision != revision {
			return conflict()
		}
		res, e := tx.ExecContext(ctx, "DELETE FROM fee_records WHERE id=? AND table_id=?", id, tableID)
		if e != nil {
			return e
		}
		n, e := res.RowsAffected()
		if e != nil {
			return e
		}
		if n == 0 {
			return missing()
		}
		return touch(ctx, tx, tableID)
	})
}
func (s *Store) Suggestions(ctx context.Context, kind string) ([]Suggestion, error) {
	if kind != "locations" && kind != "tags" {
		return nil, invalid("词条类型无效")
	}
	out := []Suggestion{}
	rows, e := s.db.QueryContext(ctx, "SELECT id,name,selection_count FROM "+kind+" ORDER BY selection_count DESC,name")
	if e != nil {
		return out, e
	}
	defer rows.Close()
	for rows.Next() {
		var v Suggestion
		if e = rows.Scan(&v.ID, &v.Name, &v.SelectionCount); e != nil {
			return out, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) AddSuggestion(ctx context.Context, kind, name string) error {
	if kind != "locations" && kind != "tags" {
		return invalid("词条类型无效")
	}
	name, e := cleanName(name)
	if e != nil {
		return e
	}
	_, e = s.db.ExecContext(ctx, "INSERT INTO "+kind+"(name) VALUES(?) ON CONFLICT(name) DO NOTHING", name)
	return e
}
