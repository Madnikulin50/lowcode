package cast2

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/spf13/cast"
)

func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

func ToInt64E(i interface{}) (int64, error) {
	i = indirect(i)

	switch s := i.(type) {
	case int:
		return int64(s), nil
	case int64:
		return s, nil
	case int32:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case uint:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	case float64:
		return int64(s), nil
	case float32:
		return int64(s), nil
	case string:
		v, err := strconv.ParseInt(s, 0, 0)
		if err == nil {
			return v, nil
		}
		v2, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return int64(v2), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", i, i)
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", i, i)
	}
}

func Uint64(in any, out *uint64) error {
	aux, err := cast.ToUint64E(in)
	if err != nil {
		return err
	}

	*out = aux
	return nil
}

func Uint(in any, out *uint) error {
	aux, err := cast.ToUintE(in)
	if err != nil {
		return err
	}

	*out = aux
	return nil
}

func Int(in any, out *int) error {
	aux, err := cast.ToIntE(in)
	if err != nil {
		return err
	}

	*out = aux
	return nil
}
