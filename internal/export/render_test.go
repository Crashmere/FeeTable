package export

import (
	"bytes"
	"feetable/internal/feetable"
	"github.com/xuri/excelize/v2"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func syntheticReport() feetable.Report {
	price := "1"
	return feetable.Report{Table: feetable.Table{ID: 1, Year: 2026, Month: 9, Revision: 1}, Records: []feetable.Record{{ID: 1, Day: 3, Location1: "合成测试起点", Location2: "用于验证中文长文本自动换行的测试目的地", Quantity: "0.145", UnitPrice: &price, Amount: "0.15"}, {ID: 2, Day: 4, Location1: "=测试文字", Location2: "固定费用", Quantity: "2.000", Amount: "100.01"}}, Count: 2, Total: "100.16"}
}
func TestFormats(t *testing.T) {
	r := syntheticReport()
	p, e := PNG(r)
	if e != nil {
		t.Fatal(e)
	}
	config, e := png.DecodeConfig(bytes.NewReader(p))
	if e != nil || config.Width != 944 || config.Height < 320 {
		t.Fatal(config, e)
	}
	pdf, e := PDF(r)
	if e != nil || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatal(e)
	}
	if !bytes.Contains(pdf, []byte("/ToUnicode")) {
		t.Fatal("PDF lacks unicode map")
	}
	x, e := XLSX(r)
	if e != nil {
		t.Fatal(e)
	}
	book, e := excelize.OpenReader(bytes.NewReader(x))
	if e != nil {
		t.Fatal(e)
	}
	defer book.Close()
	for cell, want := range map[string]string{"A1": "运费明细表", "E4": "0.145", "G4": "0.15", "F5": "", "G6": "100.16", "C5": "=测试文字"} {
		got, e := book.GetCellValue("运费明细表", cell)
		if e != nil || got != want {
			t.Fatalf("%s got %q want %q %v", cell, got, want, e)
		}
	}
	formula, e := book.GetCellFormula("运费明细表", "C5")
	if e != nil || formula != "" {
		t.Fatal("text became formula")
	}
	merges, e := book.GetMergeCells("运费明细表")
	if e != nil || len(merges) != 7 {
		t.Fatal(merges, e)
	}
	if dir := os.Getenv("FEETABLE_TEST_OUTPUT"); dir != "" {
		if e = os.MkdirAll(dir, 0700); e != nil {
			t.Fatal(e)
		}
		for name, b := range map[string][]byte{"sample.png": p, "sample.pdf": pdf, "sample.xlsx": x} {
			if e = os.WriteFile(filepath.Join(dir, name), b, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
}
func TestPNGLimit(t *testing.T) {
	r := syntheticReport()
	row := r.Records[0]
	r.Records = make([]feetable.Record, 1000)
	for i := range r.Records {
		r.Records[i] = row
	}
	if _, e := PNG(r); e == nil {
		t.Fatal("unbounded image accepted")
	}
}
