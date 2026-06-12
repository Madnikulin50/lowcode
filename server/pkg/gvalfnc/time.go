package gvalfnc

import (
	"errors"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/spf13/cast"
)

func Now() time.Time {
	return time.Now()
}

func Quarter(in any) (int, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return 0, err
	}

	return int(t.Month() / 4), nil
}

func Year(in any) (int, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return 0, err
	}

	return t.Year(), nil
}

func Month(in any) (int, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return 0, err
	}

	return int(t.Month()), nil
}

func Date(in any) (time.Time, error) {
	t, _, err := PrepMod(in, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), err
}

func Day(in any) (int, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return 0, err
	}

	return t.Day(), nil
}

func DayOf(in any) (time.Time, error) {
	return Date(in)
}
func WeekOf(in any) (time.Time, error) {
	t, _, err := PrepMod(in, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).
		AddDate(0, 0, -int(t.Weekday())), err
}

func MonthOf(in any) (time.Time, error) {
	t, _, err := PrepMod(in, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), err
}

func ThisWeek(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	ty, tw := t.ISOWeek()
	ny, nw := time.Now().ISOWeek()
	return ty == ny && tw == nw, nil
}

func PrevWeek(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	ty, tw := t.ISOWeek()
	ny, nw := time.Now().AddDate(0, 0, -7).ISOWeek()
	return ty == ny && tw == nw, nil
}

func PrevWeekTruncated(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	ty, tw := t.ISOWeek()
	ny, nw := time.Now().AddDate(0, 0, -7).ISOWeek()
	if ty == ny && tw == nw {
		return t.Weekday() <= time.Now().Weekday(), nil
	}
	return false, nil
}

func ThisYear(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now()
	return t.Year() == now.Year(), nil
}

func PrevYear(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now()
	return t.Year() == now.Year()-1, nil
}
func PrevYearTruncated(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now()
	return t.Year() == now.Year()-1 && t.YearDay() <= now.YearDay(), nil
}

func PrevMonthTruncated(in any) (bool, error) {
	p, err := PrevMonth(in)
	if err != nil {
		return false, err
	}
	if !p {
		return false, nil
	}
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now()
	return t.Day() <= now.Day(), nil
}

func PrevMonth(in any) (bool, error) {
	prevMonth := time.Now().AddDate(0, -1, 0)

	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	return t.Month() == prevMonth.Month() && t.Year() == prevMonth.Year(), nil
}

func ThisMonth(in any) (bool, error) {
	prevMonth := time.Now()

	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	return t.Month() == prevMonth.Month() && t.Year() == prevMonth.Year(), nil
}

func _quarter(t time.Time) int {
	return (int(t.Month())-1)/3 + 1
}

func ThisQuarter(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now()
	nq := _quarter(now)
	tq := _quarter(*t)
	return nq == tq && t.Year() == now.Year(), nil
}

func PrevQuarter(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now().AddDate(0, -3, 0)
	nq := _quarter(now)
	tq := _quarter(*t)
	return nq == tq && t.Year() == now.Year(), nil
}

func DayOfQuarter(t time.Time) int {
	// 1. Get the starting month of the quarter (Jan, Apr, Jul, Oct)
	quarterStartMonth := time.Month(((int(t.Month())-1)/3)*3 + 1)

	// 2. Establish the exact start date of this quarter
	quarterStart := time.Date(t.Year(), quarterStartMonth, 1, 0, 0, 0, 0, t.Location())

	// 3. Measure days between the start date and the target date
	// Subtraction yields a duration, which we convert to days, adding 1 for a 1-based index
	days := int(t.Sub(quarterStart).Hours()/24) + 1
	return days
}

func PrevQuarterTruncated(in any) (bool, error) {
	t, _, err := PrepMod(in, 0)
	if err != nil {
		return false, err
	}
	now := time.Now().AddDate(0, -3, 0)
	nq := _quarter(now)
	tq := _quarter(*t)

	return nq == tq && t.Year() == now.Year() && DayOfQuarter(*t) <= DayOfQuarter(time.Now()), nil
}

func PrepMod(base interface{}, mod interface{}) (*time.Time, int, error) {
	var (
		t *time.Time
	)

	switch auxt := base.(type) {
	case time.Time:
		t = &auxt
	case *time.Time:
		t = auxt
	case string:
		tt, err := cast.ToTimeE(auxt)

		if err != nil {
			return nil, 0, err
		}

		t = &tt
	default:
		return nil, 0, errors.New("unexpected input type")
	}

	m, err := cast.ToIntE(mod)
	if err != nil {
		return nil, 0, err
	}

	return t, m, nil
}

// Strftime formats time with POSIX standard format
// More details here:
// https://github.com/lestrrat-go/strftime#supported-conversion-specifications
func StrfTime(base interface{}, f string) (string, error) {
	t, _, err := PrepMod(base, 0)

	if err != nil {
		return "", err
	}

	o, _ := strftime.Format(f, *t,
		strftime.WithMilliseconds('b'),
		strftime.WithUnixSeconds('L'))

	return o, nil
}
