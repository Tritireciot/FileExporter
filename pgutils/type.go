package pgutils

import "fmt"

type NullString struct {
	Str   string
	Valid bool
}

type NullInteger64 struct {
	Int64 int64
	Valid bool
}

func (ns *NullString) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		ns.Str = v
		ns.Valid = true
	case []byte:
		ns.Str = string(v)
		ns.Valid = true
	case nil:
		ns.Valid = false
	default:
		ns.Valid = false
		return fmt.Errorf("incorrect type")
	}
	return nil
}

func (ns *NullInteger64) Scan(src interface{}) error {
	switch v := src.(type) {
	case int64:
		ns.Int64 = v
		ns.Valid = true
	case nil:
		ns.Int64 = 0
		ns.Valid = false
	default:
		ns.Int64 = 0
		ns.Valid = false
		return fmt.Errorf("incorrect type")
	}
	return nil
}
