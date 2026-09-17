package feetable

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var domain *Error
	if !errors.As(err, &domain) || domain.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}

func TestTableSortAndMonthPagination(t *testing.T) {
	ctx := context.Background()
	s, _ := fixture(t)
	for i := 0; i < 34; i++ {
		if _, err := s.CreateTable(ctx, TableInput{Year: 2024, Month: 2}); err != nil {
			t.Fatal(err)
		}
	}
	for _, in := range []TableInput{{Year: 2023, Month: 12}, {Year: 2025, Month: 1}} {
		if _, err := s.CreateTable(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.db.Exec("UPDATE fee_tables SET updated_at=1000-id"); err != nil {
		t.Fatal(err)
	}
	for sort, want := range map[string]int64{"": 1, "updated_desc": 1, "updated_asc": 37, "month_desc": 37, "month_asc": 36} {
		list, err := s.Tables(ctx, TableQuery{Page: 1, Sort: sort})
		if err != nil || list.Total != 37 || len(list.Items) != 30 || list.Items[0].ID != want {
			t.Fatal(sort, list, err)
		}
	}
	list, err := s.Tables(ctx, TableQuery{Page: 2, Year: 2024, Month: 2, Sort: "month_desc"})
	if err != nil || list.Total != 35 || len(list.Items) != 5 || list.Items[0].ID != 31 {
		t.Fatal(list, err)
	}
	for _, in := range []TableQuery{{Page: 0}, {Page: 1, Sort: "year; DROP TABLE fee_tables"}, {Page: 1, Year: 2024}, {Page: 1, Month: 2}} {
		_, err := s.Tables(ctx, in)
		requireCode(t, err, "VALIDATION")
	}
}

func TestLocationManagementPreservesRecords(t *testing.T) {
	ctx := context.Background()
	s, table := fixture(t)
	if _, err := s.SaveRecord(ctx, table.ID, 0, input(1)); err != nil {
		t.Fatal(err)
	}
	before, err := s.Report(ctx, table.ID, nil, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	locations, err := s.Suggestions(ctx, "locations")
	if err != nil {
		t.Fatal(err)
	}
	location := locations[0]
	for _, name := range []string{"", "换\n行", strings.Repeat("字", 81), locations[1].Name} {
		err := s.UpdateLocation(ctx, location.ID, LocationInput{Name: name, PreviousName: location.Name})
		requireCode(t, err, "VALIDATION")
	}
	if err := s.UpdateLocation(ctx, location.ID, LocationInput{Name: " 新地点 / #? ", PreviousName: location.Name}); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := s.db.QueryRow("SELECT selection_count FROM locations WHERE id=?", location.ID).Scan(&count); err != nil || count != location.SelectionCount {
		t.Fatal(count, err)
	}
	requireCode(t, s.UpdateLocation(ctx, location.ID, LocationInput{Name: "再次修改", PreviousName: location.Name}), "CONFLICT")
	requireCode(t, s.DeleteLocation(ctx, location.ID, location.Name), "CONFLICT")
	if err := s.DeleteLocation(ctx, location.ID, "新地点 / #?"); err != nil {
		t.Fatal(err)
	}
	requireCode(t, s.DeleteLocation(ctx, location.ID, "新地点 / #?"), "NOT_FOUND")
	after, err := s.Report(ctx, table.ID, nil, 1, true)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("location edit changed records", after, err)
	}
}

func mergeFixture(t *testing.T) (*Store, []Report, MergeInput) {
	t.Helper()
	ctx := context.Background()
	s, first := fixture(t)
	reports := []Report{}
	for i := 0; i < 3; i++ {
		table := first
		if i > 0 {
			var err error
			table, err = s.CreateTable(ctx, TableInput{Year: first.Year, Month: first.Month})
			if err != nil {
				t.Fatal(err)
			}
		}
		in := input(table.Revision)
		if i == 2 {
			in.Day = 1
			in.UnitPrice = nil
			in.Quantity = "-2"
			in.Amount = "-100.01"
			in.Tag = nil
		}
		if _, err := s.SaveRecord(ctx, table.ID, 0, in); err != nil {
			t.Fatal(err)
		}
		r, err := s.Report(ctx, table.ID, nil, 1, true)
		if err != nil {
			t.Fatal(err)
		}
		reports = append(reports, r)
	}
	return s, reports, MergeInput{Revision: reports[0].Table.Revision, Sources: []TableVersion{{ID: reports[1].Table.ID, Revision: reports[1].Table.Revision}, {ID: reports[2].Table.ID, Revision: reports[2].Table.Revision}}}
}

func TestMergePreservesAllRecordsAndVersions(t *testing.T) {
	ctx := context.Background()
	s, reports, in := mergeFixture(t)
	places, _ := s.Suggestions(ctx, "locations")
	target := reports[0].Table
	merged, err := s.MergeTables(ctx, target.ID, in)
	if err != nil || merged.ID != target.ID || merged.RecordCount != 3 || merged.Total != "-99.71" || merged.Revision != target.Revision+1 || merged.CreatedAt != target.CreatedAt || merged.UpdatedAt < target.UpdatedAt {
		t.Fatal(merged, err)
	}
	report, err := s.Report(ctx, target.ID, nil, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	for i, index := range []int{2, 0, 1} {
		want := reports[index].Records[0]
		want.TableID = target.ID
		if !reflect.DeepEqual(report.Records[i], want) {
			t.Fatal("record changed", report.Records[i], want)
		}
	}
	for _, source := range in.Sources {
		_, err = s.Report(ctx, source.ID, nil, 1, true)
		requireCode(t, err, "NOT_FOUND")
	}
	afterPlaces, _ := s.Suggestions(ctx, "locations")
	if !reflect.DeepEqual(places, afterPlaces) {
		t.Fatal("merge changed suggestion counts")
	}
	filtered, err := s.Report(ctx, target.ID, reports[0].Records[0].Tag, 1, true)
	if err != nil || filtered.Count != 2 || filtered.Total != "0.30" {
		t.Fatal(filtered, err)
	}
	_, err = s.SaveRecord(ctx, target.ID, reports[0].Records[0].ID, input(target.Revision))
	requireCode(t, err, "CONFLICT")
	_, err = s.MergeTables(ctx, target.ID, in)
	requireCode(t, err, "CONFLICT")
}

func TestMergeFailureIsAtomic(t *testing.T) {
	for _, kind := range []string{"target-version", "source-version", "different-year", "different-month", "missing", "duplicate", "self", "empty", "database-failure"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, reports, in := mergeFixture(t)
			switch kind {
			case "target-version":
				in.Revision--
			case "source-version":
				in.Sources[1].Revision--
			case "different-year":
				_, err := s.UpdateTable(ctx, in.Sources[1].ID, TableInput{Year: 2023, Month: 2, Revision: in.Sources[1].Revision})
				if err != nil {
					t.Fatal(err)
				}
				in.Sources[1].Revision++
			case "different-month":
				_, err := s.UpdateTable(ctx, in.Sources[1].ID, TableInput{Year: 2024, Month: 3, Revision: in.Sources[1].Revision})
				if err != nil {
					t.Fatal(err)
				}
				in.Sources[1].Revision++
			case "missing":
				in.Sources[1].ID = 999
			case "duplicate":
				in.Sources[1] = in.Sources[0]
			case "self":
				in.Sources[1].ID = reports[0].Table.ID
			case "empty":
				in.Sources = nil
			case "database-failure":
				_, err := s.db.Exec(fmt.Sprintf("CREATE TRIGGER fail_merge BEFORE DELETE ON fee_tables WHEN OLD.id=%d BEGIN SELECT RAISE(ABORT,'injected'); END", in.Sources[1].ID))
				if err != nil {
					t.Fatal(err)
				}
			}
			for i := range reports {
				var err error
				reports[i], err = s.Report(ctx, reports[i].Table.ID, nil, 1, true)
				if err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.MergeTables(ctx, reports[0].Table.ID, in); err == nil {
				t.Fatal("invalid merge succeeded")
			}
			for _, before := range reports {
				after, err := s.Report(ctx, before.Table.ID, nil, 1, true)
				if err != nil || !reflect.DeepEqual(before, after) {
					t.Fatal("partial merge", before, after, err)
				}
			}
		})
	}
}

func TestMergeRecordLimitAndEmptyTables(t *testing.T) {
	ctx := context.Background()
	s, first := fixture(t)
	second, err := s.CreateTable(ctx, TableInput{Year: first.Year, Month: first.Month})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec("WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<5000) INSERT INTO fee_records(table_id,day,location1,location2,quantity_milli,amount_cents) SELECT ?,1,'起点','终点',1,1 FROM n", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.SaveRecord(ctx, second.ID, 0, input(1)); err != nil {
		t.Fatal(err)
	}
	in := MergeInput{Revision: 1, Sources: []TableVersion{{ID: second.ID, Revision: 2}}}
	_, err = s.MergeTables(ctx, first.ID, in)
	requireCode(t, err, "VALIDATION")
	if _, err = s.db.Exec("DELETE FROM fee_records WHERE table_id=?", second.ID); err != nil {
		t.Fatal(err)
	}
	merged, err := s.MergeTables(ctx, first.ID, in)
	if err != nil || merged.RecordCount != 5000 || merged.Total != "50.00" {
		t.Fatal(merged, err)
	}
}
