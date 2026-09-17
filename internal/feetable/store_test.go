package feetable

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) (*Store, Table) {
	t.Helper()
	s, e := Open(filepath.Join(t.TempDir(), "test.sqlite"), true)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	table, e := s.CreateTable(context.Background(), TableInput{Year: 2024, Month: 2})
	if e != nil {
		t.Fatal(e)
	}
	return s, table
}
func input(revision int64) RecordInput {
	p := "1"
	tag := "客户/甲 #1"
	return RecordInput{Day: 29, Location1: "测试起点", Location2: "测试终点", Quantity: "0.145", UnitPrice: &p, Tag: &tag, Revision: revision}
}
func TestExactMoney(t *testing.T) {
	for _, tc := range []struct{ q, p, want string }{{"0.145", "1", "0.15"}, {"0.285", "1", "0.29"}, {"-1.025", "1", "-1.03"}, {"2.125", "40", "85.00"}, {"3.000", "0", "0.00"}} {
		q, e := ParseDecimal(tc.q, 3)
		if e != nil {
			t.Fatal(e)
		}
		p, _ := ParseDecimal(tc.p, 2)
		a, e := Calculate(q, p)
		if e != nil || FormatDecimal(a, 2) != tc.want {
			t.Fatalf("%+v got %d %v", tc, a, e)
		}
	}
	for _, v := range []string{"NaN", "Infinity", "1e2", "1,23", "1.2345", "", "-"} {
		if _, e := ParseDecimal(v, 3); e == nil {
			t.Fatalf("accepted %q", v)
		}
	}
	if _, e := Calculate(MaxValue, MaxValue); e == nil {
		t.Fatal("overflow accepted")
	}
}
func TestRecordLifecycle(t *testing.T) {
	ctx := context.Background()
	s, table := fixture(t)
	r, e := s.SaveRecord(ctx, table.ID, 0, input(table.Revision))
	if e != nil || r.Amount != "0.15" {
		t.Fatalf("save %+v %v", r, e)
	}
	report, e := s.Report(ctx, table.ID, r.Tag, 1, true)
	if e != nil || report.Total != "0.15" || report.Count != 1 {
		t.Fatal(report, e)
	}
	stale := input(table.Revision)
	if _, e = s.SaveRecord(ctx, table.ID, 0, stale); e == nil {
		t.Fatal("stale write accepted")
	}
	fixed := input(report.Table.Revision)
	fixed.Day = 28
	fixed.UnitPrice = nil
	fixed.Amount = "100.01"
	fixed.Tag = nil
	fixed.Quantity = "1"
	fixedRecord, e := s.SaveRecord(ctx, table.ID, 0, fixed)
	if e != nil || fixedRecord.UnitPrice != nil || fixedRecord.Amount != "100.01" {
		t.Fatal(fixedRecord, e)
	}
	full, _ := s.Report(ctx, table.ID, nil, 1, true)
	if full.Total != "100.16" || full.Records[0].ID != fixedRecord.ID {
		t.Fatal(full)
	}
	if _, e = s.UpdateTable(ctx, table.ID, TableInput{Year: 2025, Month: 2, Revision: full.Table.Revision}); e == nil {
		t.Fatal("invalid leap date accepted")
	}
	bad := input(full.Table.Revision)
	bad.Day = 31
	if _, e = s.SaveRecord(ctx, table.ID, 0, bad); e == nil {
		t.Fatal("invalid date accepted")
	}
	after, _ := s.Report(ctx, table.ID, nil, 1, true)
	if after.Table.Revision != full.Table.Revision || after.Count != 2 {
		t.Fatal("failed write changed state")
	}
	edited := input(full.Table.Revision)
	edited.Location1 = "新的地点"
	if _, e = s.SaveRecord(ctx, table.ID, r.ID, edited); e != nil {
		t.Fatal(e)
	}
	places, _ := s.Suggestions(ctx, "locations")
	found := false
	for _, p := range places {
		if p.Name == "新的地点" {
			found = true
			if p.SelectionCount != 0 {
				t.Fatal(p)
			}
		}
	}
	if !found {
		t.Fatal("edited place missing")
	}
	current, _ := s.Report(ctx, table.ID, nil, 1, true)
	if e = s.DeleteTable(ctx, table.ID, current.Table.Revision); e != nil {
		t.Fatal(e)
	}
	var n int
	s.db.QueryRow("SELECT COUNT(*) FROM fee_records").Scan(&n)
	if n != 0 {
		t.Fatal("cascade failed")
	}
}
func TestBackupRestoreAndIdentity(t *testing.T) {
	ctx := context.Background()
	s, table := fixture(t)
	if _, e := s.SaveRecord(ctx, table.ID, 0, input(table.Revision)); e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	backup := filepath.Join(dir, "backup.sqlite")
	if e := s.Backup(ctx, backup); e != nil {
		t.Fatal(e)
	}
	if e := s.Backup(ctx, backup); e == nil {
		t.Fatal("overwrote backup")
	}
	restored := filepath.Join(dir, "restored.sqlite")
	if e := Restore(ctx, backup, restored); e != nil {
		t.Fatal(e)
	}
	r, e := Open(restored, false)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	report, e := r.Report(ctx, table.ID, nil, 1, true)
	if e != nil || report.Total != "0.15" {
		t.Fatal(report, e)
	}
	if _, e = Open(filepath.Join(dir, "missing.sqlite"), false); e == nil {
		t.Fatal("missing DB created")
	}
	if e = CheckFile(ctx, filepath.Join(dir, "missing.sqlite")); e == nil {
		t.Fatal("missing check succeeded")
	}
}
func TestDuplicateMonthsAndPagination(t *testing.T) {
	ctx := context.Background()
	s, table := fixture(t)
	if _, e := s.CreateTable(ctx, TableInput{Year: table.Year, Month: table.Month}); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 53; i++ {
		in := input(int64(i + 1))
		in.Day = 1
		if _, e := s.SaveRecord(ctx, table.ID, 0, in); e != nil {
			t.Fatal(e)
		}
	}
	page, e := s.Report(ctx, table.ID, nil, 2, false)
	if e != nil || len(page.Records) != 3 || page.Count != 53 || page.Total != "7.95" {
		t.Fatal(page, e)
	}
	all, _ := s.Report(ctx, table.ID, nil, 1, true)
	if len(all.Records) != 53 {
		t.Fatal("export lost rows")
	}
	_, e = s.Report(ctx, 999, nil, 1, true)
	var domain *Error
	if !errors.As(e, &domain) || domain.Code != "NOT_FOUND" {
		t.Fatal(e)
	}
}
