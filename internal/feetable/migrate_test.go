package feetable

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDecimalPriceMigrationPreservesData(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "v1.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	v1 := strings.Replace(schema, "unit_price_cents>=0)", "(unit_price_cents>=0 AND unit_price_cents%100=0))", 1)
	v1 = strings.Replace(v1, "user_version=2", "user_version=1", 1)
	_, err = db.Exec(v1 + `
INSERT INTO fee_tables(id,year,month,created_at,updated_at,revision) VALUES(7,2026,9,1000,2000,9);
INSERT INTO fee_records VALUES(21,7,1,'起点','终点',10500,200,2100,'甲');
INSERT INTO fee_records VALUES(22,7,2,'起点','终点',1000,NULL,999,NULL);
INSERT INTO fee_records VALUES(90,7,3,'起点','终点',0,0,0,NULL);
DELETE FROM fee_records WHERE id=90;
INSERT INTO locations VALUES(3,'起点',4);
INSERT INTO tags VALUES(5,'甲',8);
`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE fee_records SET unit_price_cents=288 WHERE id=21"); err == nil {
		t.Fatal("v1 must reject decimal prices")
	}
	db.Close()
	s, err := Open(path, false)
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.Report(ctx, 7, nil, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if s.RequireCurrentSchema(ctx) == nil {
		t.Fatal("v1 allowed for serving")
	}
	backup := filepath.Join(t.TempDir(), "before.sqlite")
	if err := s.Backup(ctx, backup); err != nil {
		t.Fatal(err)
	}
	s.Close()
	for i := 0; i < 2; i++ {
		if err := Migrate(ctx, path); err != nil {
			t.Fatal(err)
		}
	}
	s, err = Open(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.RequireCurrentSchema(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := s.Report(ctx, 7, nil, 1, true)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("migration changed records", before, after, err)
	}
	for _, kind := range []string{"locations", "tags"} {
		values, err := s.Suggestions(ctx, kind)
		if err != nil || len(values) != 1 || values[0].SelectionCount == 0 {
			t.Fatal(values, err)
		}
	}
	price := "2.88"
	r, err := s.SaveRecord(ctx, 7, 0, RecordInput{Day: 1, Location1: "起点", Location2: "终点", Quantity: "10.5", UnitPrice: &price, Revision: 9})
	if err != nil || r.ID <= 90 || r.Amount != "30.24" || *r.UnitPrice != price {
		t.Fatal(r, err)
	}
	if err := s.DeleteTable(ctx, 7, 10); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM fee_records").Scan(&count); err != nil || count != 0 {
		t.Fatal("cascade lost", count, err)
	}
	old, err := Open(backup, false)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	preserved, err := old.Report(ctx, 7, nil, 1, true)
	if err != nil || !reflect.DeepEqual(before, preserved) {
		t.Fatal("backup changed", preserved, err)
	}
}
