package gvalfnc

import (
	"math"

	"github.com/modern-go/reflect2"
	"github.com/spf13/cast"
)

func IsNil(i any) bool {
	return reflect2.IsNil(i)
}

func CastFloat(i any) (float64, error) {
	return cast.ToFloat64E(i)
}

func Anomality(val any, avg any, stddev any) (float64, error) {
	v, err := CastFloat(val)
	if err != nil {
		return 0, err
	}
	a, err := CastFloat(avg)
	if err != nil {
		return 0, err
	}
	sd, err := CastFloat(stddev)
	if err != nil {
		return 0, err
	}
	t := math.Abs(v - a)
	if t > sd {
		return t / sd, nil
	}

	return 0, nil
}

func CastInt(i any) (int, error) {
	return cast.ToIntE(i)
}
func CastString(i any) (string, error) {
	return cast.ToStringE(i)
}
