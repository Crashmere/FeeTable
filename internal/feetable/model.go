package feetable

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string     { return e.Message }
func invalid(message string) error { return &Error{"VALIDATION", message} }
func missing() error               { return &Error{"NOT_FOUND", "记录不存在或已被删除"} }
func conflict() error {
	return &Error{"CONFLICT", "这张表已被修改，请刷新后核对再保存"}
}

const MaxRecords = 5000
const MaxValue int64 = 1_000_000_000_000

type Table struct {
	ID          int64  `json:"id"`
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Revision    int64  `json:"revision"`
	RecordCount int    `json:"recordCount"`
	Total       string `json:"total"`
}
type Record struct {
	ID        int64   `json:"id"`
	TableID   int64   `json:"tableId"`
	Day       int     `json:"day"`
	Location1 string  `json:"location1"`
	Location2 string  `json:"location2"`
	Quantity  string  `json:"quantity"`
	UnitPrice *string `json:"unitPrice"`
	Amount    string  `json:"amount"`
	Tag       *string `json:"tag"`
}
type TableInput struct {
	Year     int   `json:"year"`
	Month    int   `json:"month"`
	Revision int64 `json:"revision"`
}
type TableQuery struct {
	Page        int
	Sort        string
	Year, Month int
}
type TableVersion struct {
	ID       int64 `json:"id"`
	Revision int64 `json:"revision"`
}
type MergeInput struct {
	Revision int64          `json:"revision"`
	Sources  []TableVersion `json:"sources"`
}
type LocationInput struct {
	Name         string `json:"name"`
	PreviousName string `json:"previousName"`
}
type RecordInput struct {
	Day       int     `json:"day"`
	Location1 string  `json:"location1"`
	Location2 string  `json:"location2"`
	Quantity  string  `json:"quantity"`
	UnitPrice *string `json:"unitPrice"`
	Amount    string  `json:"amount"`
	Tag       *string `json:"tag"`
	Revision  int64   `json:"revision"`
}
type Suggestion struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	SelectionCount int64  `json:"selectionCount"`
}
type Report struct {
	Table     Table    `json:"table"`
	Records   []Record `json:"records"`
	Tags      []string `json:"tags"`
	FilterTag *string  `json:"filterTag"`
	Count     int      `json:"count"`
	Total     string   `json:"total"`
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
}
type TableList struct {
	Items    []Table `json:"items"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
}

// ParseDecimal accepts plain base-ten input and never passes through binary floating point.
func ParseDecimal(value string, scale int) (int64, error) {
	value = strings.TrimSpace(value)
	negative := false
	if strings.HasPrefix(value, "-") {
		negative = true
		value = value[1:]
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(value) > 24 {
		return 0, invalid("请输入有效的十进制数字")
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" {
			return 0, invalid("小数点后需要数字")
		}
	}
	if len(fraction) > scale {
		return 0, invalid(fmt.Sprintf("最多允许 %d 位小数", scale))
	}
	digits := parts[0] + fraction + strings.Repeat("0", scale-len(fraction))
	for _, r := range digits {
		if r < '0' || r > '9' {
			return 0, invalid("请输入有效的十进制数字")
		}
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || n > MaxValue {
		return 0, invalid("数值超出允许范围")
	}
	if negative {
		n = -n
	}
	return n, nil
}
func FormatDecimal(n int64, scale int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	base := int64(1)
	for i := 0; i < scale; i++ {
		base *= 10
	}
	if scale == 0 {
		return sign + strconv.FormatInt(n, 10)
	}
	return fmt.Sprintf("%s%d.%0*d", sign, n/base, scale, n%base)
}
func Calculate(quantity int64, price int64) (int64, error) {
	n := new(big.Int).Mul(big.NewInt(quantity), big.NewInt(price))
	negative := n.Sign() < 0
	n.Abs(n)
	n.Add(n, big.NewInt(500))
	n.Quo(n, big.NewInt(1000))
	if negative {
		n.Neg(n)
	}
	if !n.IsInt64() || n.Cmp(big.NewInt(MaxValue)) > 0 || n.Cmp(big.NewInt(-MaxValue)) < 0 {
		return 0, invalid("金额超出允许范围")
	}
	return n.Int64(), nil
}
func validMonth(year, month int) error {
	if year < 1 || year > 9999 || month < 1 || month > 12 {
		return invalid("年份须为 1–9999，月份须为 1–12")
	}
	return nil
}
func validDay(year, month, day int) error {
	if day < 1 || day > time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day() {
		return invalid("日期不在当前月份内")
	}
	return nil
}
func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return "", invalid("名称须为 1–80 个字符")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", invalid("名称不能包含换行或控制字符")
		}
	}
	return name, nil
}
