package export

import (
	"bytes"
	"feetable/internal/feetable"
	"github.com/xuri/excelize/v2"
	"image/png"
	"os"
	"path/filepath"
	"strings"
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

func TestLongNamesAndLargeTotalRemainVisible(t *testing.T) {
	r := syntheticReport()
	r.Records[0].Location1 = strings.Repeat("汉", 80)
	r.Total = "50000000000000.00"
	layout, err := BuildLayout(r)
	if err != nil {
		t.Fatal(err)
	}
	footer := layout.Cells[len(layout.Cells)-1]
	if footer.H < 68 || int(footer.Y+footer.H) > layout.Height-32 {
		t.Fatal("large total clipped", footer, layout.Height)
	}
	data, err := XLSX(r)
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	height, err := book.GetRowHeight("运费明细表", 4)
	if err != nil || height < 120 {
		t.Fatal("long name row too short", height, err)
	}
}

func TestDecimalPriceExports(t *testing.T) {
	r := syntheticReport()
	price := "2.88"
	r.Records = r.Records[:1]
	r.Records[0].Quantity, r.Records[0].UnitPrice, r.Records[0].Amount = "10.500", &price, "30.24"
	r.Total, r.Count = "30.24", 1
	layout, err := BuildLayout(r)
	if err != nil || layout.Cells[15].Text != "2.88" || layout.Cells[16].Text != "30.24" {
		t.Fatal(layout, err)
	}
	if _, err := PNG(r); err != nil {
		t.Fatal(err)
	}
	if _, err := PDF(r); err != nil {
		t.Fatal(err)
	}
	data, err := XLSX(r)
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	for cell, want := range map[string]string{"E4": "10.500", "F4": "2.88", "G4": "30.24", "G5": "30.24"} {
		got, err := book.GetCellValue("运费明细表", cell)
		if err != nil || got != want {
			t.Fatalf("%s: got %q want %q (%v)", cell, got, want, err)
		}
	}
}
