package transformer

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

const maxExactFloatInteger = 1 << 53

func finiteNumber(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("арифметика не поддерживает NaN и бесконечность")
	}
	return value, nil
}

func exactSigned(value int64) (float64, error) {
	if value < -maxExactFloatInteger || value > maxExactFloatInteger {
		return 0, fmt.Errorf("целое число %d нельзя точно представить как float64", value)
	}
	return float64(value), nil
}


func exactUnsigned(value uint64) (float64, error) {
	if value > maxExactFloatInteger {
		return 0, fmt.Errorf("целое число %d нельзя точно представить как float64", value)
	}
	return float64(value), nil
}

func toNumber(value any) (float64, error) {
	if number, ok := value.(json.Number); ok  {
				text := string(number)
		if !strings.ContainsAny(text, ".eE") {
			if signed, err := strconv.ParseInt(text, 10, 64); err == nil {
				return exactSigned(signed)
			}
			if unsigned, err := strconv.ParseUint(text, 10, 64); err == nil {
				return exactUnsigned(unsigned)
			}
			return 0, fmt.Errorf("целое число %q выходит за поддерживаемый диапазон", text)
		}

		parsed, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, fmt.Errorf("некорректное число %q: %w", text, err)
		}
		return finiteNumber(parsed)
	}

	if value == nil {
		return 0, fmt.Errorf("ожидалось число, получено nil")
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return exactSigned(v.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return exactUnsigned(v.Uint())
	case reflect.Float32, reflect.Float64:
		return finiteNumber(v.Float())
	default:
		return 0, fmt.Errorf("ожидалось число, получено %T", value)
	}
}

func add(a, b any) (float64, error) {
	left, err := toNumber(a)
	if err != nil {
		return 0, err
	}
	right, err := toNumber(b)
	if err != nil {
		return 0, err
	}
	return left + right, nil
}

func sub(a, b any) (float64, error) {
	left, err := toNumber(a)
	if err != nil {
		return 0, err
	}
	right, err := toNumber(b)
	if err != nil {
		return 0, err
	}
	return left - right, nil
}

func mul(a, b any) (float64, error) {
	left, err := toNumber(a)
	if err != nil {
		return 0, err
	}
	right, err := toNumber(b)
	if err != nil {
		return 0, err
	}
	return left * right, nil
}

func div(a, b any) (float64, error) {
	left, err := toNumber(a)
	if err != nil {
		return 0, err
	}
	right, err := toNumber(b)
	if err != nil {
		return 0, err
	}

	if right == 0.0 {
		return 0, fmt.Errorf("деление на ноль")
	}
	return left / right, nil
}
