package pgutils

import "fmt"

type NullString struct {
	Str   string
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