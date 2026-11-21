package cast2

import "github.com/spf13/cast"

func Float64(in any, out *float64) error {
	aux, err := cast.ToFloat64E(in)
	if err != nil {
		return err
	}

	*out = aux
	return nil
}
