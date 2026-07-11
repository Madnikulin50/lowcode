package clickhouse

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers"
)

type (
	TypeTimestamp struct{ *dal.TypeTimestamp }
	TypeBoolean   struct{ *dal.TypeBoolean }
)

func (*TypeTimestamp) MakeScanBuffer() any { return new(sql.NullTime) }

func (t *TypeTimestamp) Decode(raw any) (any, bool, error) {
	dec, is := raw.(*sql.NullTime)
	if !is {
		return nil, false, fmt.Errorf("unexpected raw type %T for Timestamp", raw)
	}

	if dec.Valid {
		return dec.Time.UTC().Format(drivers.TimestampLayout(t.Timezone, t.Precision)), dec.Valid, nil
	}

	return nil, false, nil
}

func (t *TypeTimestamp) Encode(val any) (driver.Value, error) {
	return val, nil
}

func (*TypeBoolean) MakeScanBuffer() any { return new(sql.NullInt64) }

func (t *TypeBoolean) Decode(raw any) (any, bool, error) {
	dec, is := raw.(*sql.NullInt64)
	if !is {
		return nil, false, fmt.Errorf("unexpected raw type %T for Boolean", raw)
	}

	return dec.Int64 != 0 && dec.Valid, dec.Valid, nil
}

func (t *TypeBoolean) Encode(val any) (driver.Value, error) {
	if val == nil {
		return nil, nil
	}

	if v, ok := val.(bool); ok {
		if v {
			return int64(1), nil
		}
		return int64(0), nil
	}

	return val, nil
}
