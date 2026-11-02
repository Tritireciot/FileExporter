package transformer

import (
	"fmt"
	"strings"
)

func compareValues(a, b any, allowBool bool) (int, error) {
	switch left := a.(type) {
	case string:
		right, ok := b.(string)
		if !ok {
			return 0, fmt.Errorf("нельзя сравнить строку с %T", b)
		}
		return strings.Compare(left, right), nil

	case bool:
		right, ok := b.(bool)
		if !ok {
			return 0, fmt.Errorf("нельзя сравнить bool с %T", b)
		}
		if !allowBool {
			return 0, fmt.Errorf("для bool допустимы только == и !=")
		}
		if left == right {
			return 0, nil
		}
		return 1, nil // Для != достаточно знать, что значения различаются.

	default:
		leftNumber, err := toNumber(a)
		if err != nil {
			return 0, err
		}
		rightNumber, err := toNumber(b)
		if err != nil {
			return 0, err
		}

		switch {
		case leftNumber < rightNumber:
			return -1, nil
		case leftNumber > rightNumber:
			return 1, nil
		default:
			return 0, nil
		}
	}
}

func templateEq(a, b any) (bool, error) {
	result, err := compareValues(a, b, true)
	return result == 0, err
}

func templateNe(a, b any) (bool, error) {
	res, err := templateEq(a, b)
	return !res, err
}

func templateGt(a, b any) (bool, error) {
	result, err := compareValues(a, b, false)
	return result > 0, err
}

func templateLt(a, b any) (bool, error) {
	res, err := templateGt(a, b)
	return !res, err
}

func templateGe(a, b any) (bool, error) {
	result, err := compareValues(a, b, false)
	return result >= 0, err
}

func templateLe(a, b any) (bool, error) {
	result, err := compareValues(a, b, false)
	return result <= 0, err
}