package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Decimal struct {
	value string
}

func NewDecimalFromFloat64(f float64) Decimal {
	return Decimal{value: fmt.Sprintf("%.2f", f)}
}

func NewDecimalFromString(s string) Decimal {
	return Decimal{value: s}
}

func NewDecimalFromCents(cents int64) Decimal {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return Decimal{value: fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)}
}

func (d Decimal) String() string {
	return d.value
}

func (d Decimal) Value() (driver.Value, error) {
	return d.value, nil
}

func (d *Decimal) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		d.value = "0.00"
		return nil
	case []byte:
		d.value = string(v)
		return nil
	case string:
		d.value = v
		return nil
	case float64:
		d.value = fmt.Sprintf("%.2f", v)
		return nil
	case int64:
		d.value = fmt.Sprintf("%d.00", v)
		return nil
	default:
		return errors.New("unsupported decimal type")
	}
}

func (d Decimal) ToCents() (int64, error) {
	val := strings.TrimSpace(d.value)
	if val == "" {
		return 0, nil
	}
	sign := int64(1)
	if strings.HasPrefix(val, "-") {
		sign = -1
		val = strings.TrimPrefix(val, "-")
	}
	parts := strings.SplitN(val, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	frac := "00"
	if len(parts) > 1 {
		frac = parts[1]
	}
	if len(frac) == 1 {
		frac += "0"
	} else if len(frac) > 2 {
		frac = frac[:2]
	}
	fracVal, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, err
	}
	return sign * (whole*100 + fracVal), nil
}
