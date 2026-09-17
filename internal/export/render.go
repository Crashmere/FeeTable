package export

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strconv"
	"sync"

	"feetable/internal/feetable"
	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/LXGWNeoXiHei.ttf
var fontBytes []byte
var parsedFont *opentype.Font
var fontErr error
var fontOnce sync.Once

type Cell struct {
	X, Y, W, H float64
	Text       string
	Align      string
	Size       float64
}
type Layout struct {
	Width, Height int
	Cells         []Cell
}

func face(size float64) (font.Face, error) {
	fontOnce.Do(func() { parsedFont, fontErr = opentype.Parse(fontBytes) })
	if fontErr != nil {
		return nil, fontErr
	}
	return opentype.NewFace(parsedFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}
func wrap(text string, width float64, f font.Face) []string {
	lines := []string{}
	current := ""
	for _, r := range text {
		next := current + string(r)
		if current != "" && float64(font.MeasureString(f, next))/64 > width {
			lines = append(lines, current)
			current = string(r)
		} else {
			current = next
		}
	}
	return append(lines, current)
}
func BuildLayout(report feetable.Report) (Layout, error) {
	f, e := face(18)
	if e != nil {
		return Layout{}, e
	}
	defer f.Close()
	// Fixed report geometry with wrapping keeps long names from producing unbounded images.
	widths := []float64{58, 58, 188, 188, 130, 100, 158}
	total := float64(0)
	for _, w := range widths {
		total += w
	}
	layout := Layout{Width: int(total + 64)}
	add := func(x, y, w, h float64, text, align string, size float64) {
		layout.Cells = append(layout.Cells, Cell{x, y, w, h, text, align, size})
	}
	x, y := 32.0, 32.0
	add(x, y, total, 58, "运费明细表", "center", 28)
	y += 58
	add(x, y, 116, 42, fmt.Sprintf("%d年", report.Table.Year), "center", 20)
	add(x+116, y, 376, 42, "摘要", "center", 20)
	add(x+492, y, 130, 84, "运输数量", "center", 20)
	add(x+622, y, 100, 84, "单价", "center", 20)
	add(x+722, y, 158, 84, "运输金额", "center", 20)
	y += 42
	add(x, y, 58, 42, "月", "center", 18)
	add(x+58, y, 58, 42, "日", "center", 18)
	add(x+116, y, 188, 42, "", "left", 18)
	add(x+304, y, 188, 42, "", "left", 18)
	y += 42
	for _, r := range report.Records {
		p := ""
		if r.UnitPrice != nil {
			p = *r.UnitPrice
		}
		texts := []string{strconv.Itoa(report.Table.Month), strconv.Itoa(r.Day), r.Location1, r.Location2, r.Quantity, p, r.Amount}
		height := 48.0
		for i, text := range texts {
			n := len(wrap(text, widths[i]-20, f))
			height = math.Max(height, float64(n)*24+20)
		}
		x = 32
		for i, text := range texts {
			align := "center"
			if i == 2 || i == 3 {
				align = "left"
			}
			if i == 4 || i == 6 {
				align = "right"
			}
			add(x, y, widths[i], height, text, align, 18)
			x += widths[i]
		}
		y += height
	}
	footerHeight := math.Max(48, float64(len(wrap(report.Total, 138, f)))*24+20)
	add(32, y, total-158, footerHeight, "", "left", 18)
	add(32+total-158, y, 158, footerHeight, report.Total, "right", 18)
	layout.Height = int(y + footerHeight + 32)
	return layout, nil
}
func PNG(report feetable.Report) ([]byte, error) {
	layout, e := BuildLayout(report)
	if e != nil {
		return nil, e
	}
	if int64(layout.Width)*int64(layout.Height) > 20_000_000 {
		return nil, &feetable.Error{Code: "LIMIT", Message: "图片过长，请按标签分开导出，或使用 PDF / Excel"}
	}
	img := image.NewRGBA(image.Rect(0, 0, layout.Width, layout.Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	faces := map[float64]font.Face{}
	defer func() {
		for _, f := range faces {
			f.Close()
		}
	}()
	line := func(x, y, w, h int) {
		draw.Draw(img, image.Rect(x, y, x+w, y+h), image.NewUniform(color.Black), image.Point{}, draw.Src)
	}
	for _, c := range layout.Cells {
		x, y, w, h := int(c.X), int(c.Y), int(c.W), int(c.H)
		line(x, y, w, 1)
		line(x, y+h, w+1, 1)
		line(x, y, 1, h)
		line(x+w, y, 1, h)
		f := faces[c.Size]
		if f == nil {
			f, e = face(c.Size)
			if e != nil {
				return nil, e
			}
			faces[c.Size] = f
		}
		lines := wrap(c.Text, c.W-20, f)
		lineHeight := c.Size + 6
		top := c.Y + (c.H-float64(len(lines))*lineHeight)/2
		for i, text := range lines {
			tw := float64(font.MeasureString(f, text)) / 64
			tx := c.X + 10
			if c.Align == "right" {
				tx = c.X + c.W - 10 - tw
			} else if c.Align == "center" {
				tx = c.X + (c.W-tw)/2
			}
			d := font.Drawer{Dst: img, Src: image.NewUniform(color.Black), Face: f, Dot: fixed.P(int(math.Round(tx)), int(math.Round(top+float64(i)*lineHeight+float64(f.Metrics().Ascent)/64)))}
			d.DrawString(text)
		}
	}
	var out bytes.Buffer
	e = png.Encode(&out, img)
	return out.Bytes(), e
}
func PDF(report feetable.Report) ([]byte, error) {
	layout, e := BuildLayout(report)
	if e != nil {
		return nil, e
	}
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: float64(layout.Width), H: float64(layout.Height)}})
	pdf.AddPage()
	if e = pdf.AddTTFFontData("report", fontBytes); e != nil {
		return nil, e
	}
	pdf.SetLineWidth(0.7)
	faces := map[float64]font.Face{}
	defer func() {
		for _, f := range faces {
			f.Close()
		}
	}()
	for _, c := range layout.Cells {
		pdf.RectFromUpperLeft(c.X, c.Y, c.W, c.H)
		if e = pdf.SetFont("report", "", c.Size); e != nil {
			return nil, e
		}
		f := faces[c.Size]
		if f == nil {
			f, e = face(c.Size)
			if e != nil {
				return nil, e
			}
			faces[c.Size] = f
		}
		lines := wrap(c.Text, c.W-20, f)
		lh := c.Size + 6
		y := c.Y + (c.H-float64(len(lines))*lh)/2
		alignment := gopdf.Left | gopdf.Middle
		if c.Align == "right" {
			alignment = gopdf.Right | gopdf.Middle
		} else if c.Align == "center" {
			alignment = gopdf.Center | gopdf.Middle
		}
		for i, text := range lines {
			pdf.SetXY(c.X+10, y+float64(i)*lh)
			if e = pdf.CellWithOption(&gopdf.Rect{W: c.W - 20, H: lh}, text, gopdf.CellOption{Align: alignment}); e != nil {
				return nil, e
			}
		}
	}
	var out bytes.Buffer
	_, e = pdf.WriteTo(&out)
	return out.Bytes(), e
}
func XLSX(report feetable.Report) ([]byte, error) {
	layout, e := BuildLayout(report)
	if e != nil {
		return nil, e
	}
	f := excelize.NewFile()
	defer f.Close()
	sheet := "运费明细表"
	if e := f.SetSheetName("Sheet1", sheet); e != nil {
		return nil, e
	}
	border := []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}}
	base, e := f.NewStyle(&excelize.Style{Border: border, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	if e != nil {
		return nil, e
	}
	title, e := f.NewStyle(&excelize.Style{Border: border, Font: &excelize.Font{Bold: true, Size: 16}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	if e != nil {
		return nil, e
	}
	qtyFormat, amountFormat := "0.000", "0.00"
	quantity, e := f.NewStyle(&excelize.Style{Border: border, CustomNumFmt: &qtyFormat, Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})
	if e != nil {
		return nil, e
	}
	amount, e := f.NewStyle(&excelize.Style{Border: border, CustomNumFmt: &amountFormat, Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})
	if e != nil {
		return nil, e
	}
	end := len(report.Records) + 4
	if e = f.SetCellStyle(sheet, "A1", fmt.Sprintf("G%d", end), base); e != nil {
		return nil, e
	}
	for _, m := range [][2]string{{"A1", "G1"}, {"A2", "B2"}, {"C2", "D2"}, {"E2", "E3"}, {"F2", "F3"}, {"G2", "G3"}, {fmt.Sprintf("A%d", end), fmt.Sprintf("F%d", end)}} {
		if e = f.MergeCell(sheet, m[0], m[1]); e != nil {
			return nil, e
		}
	}
	for cell, text := range map[string]string{"A1": "运费明细表", "A2": fmt.Sprintf("%d年", report.Table.Year), "C2": "摘要", "E2": "运输数量", "F2": "单价", "G2": "运输金额", "A3": "月", "B3": "日"} {
		if e = f.SetCellStr(sheet, cell, text); e != nil {
			return nil, e
		}
	}
	if e = f.SetCellStyle(sheet, "A1", "G3", title); e != nil {
		return nil, e
	}
	for i, r := range report.Records {
		row := i + 4
		q, _ := strconv.ParseFloat(r.Quantity, 64)
		a, _ := strconv.ParseFloat(r.Amount, 64)
		var p any
		if r.UnitPrice != nil {
			p, _ = strconv.ParseInt(*r.UnitPrice, 10, 64)
		}
		values := []any{report.Table.Month, r.Day, r.Location1, r.Location2, q, p, a}
		if e = f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &values); e != nil {
			return nil, e
		}
		if e = f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), quantity); e != nil {
			return nil, e
		}
		if e = f.SetCellStyle(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), amount); e != nil {
			return nil, e
		}
		// Use the wrapped report geometry so long names remain visible in Excel.
		if e = f.SetRowHeight(sheet, row, layout.Cells[10+i*7].H*0.75); e != nil {
			return nil, e
		}
	}
	total, _ := strconv.ParseFloat(report.Total, 64)
	cell := fmt.Sprintf("G%d", end)
	if e = f.SetCellValue(sheet, cell, total); e != nil {
		return nil, e
	}
	if e = f.SetCellStyle(sheet, cell, cell, amount); e != nil {
		return nil, e
	}
	for _, c := range []struct {
		a, b string
		w    float64
	}{{"A", "B", 7}, {"C", "D", 26}, {"E", "F", 16}, {"G", "G", 22}} {
		if e = f.SetColWidth(sheet, c.a, c.b, c.w); e != nil {
			return nil, e
		}
	}
	for _, row := range []int{1, 2, 3, end} {
		if e = f.SetRowHeight(sheet, row, 30); e != nil {
			return nil, e
		}
	}
	out, e := f.WriteToBuffer()
	if e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
