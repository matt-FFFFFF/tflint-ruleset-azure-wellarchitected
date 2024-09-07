package blockquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/tidwall/gjson"
)

type ResultCompareFunc func(gjson.Result, ...gjson.Result) (bool, string, error)

func Exists(r gjson.Result, _ ...gjson.Result) (bool, string, error) {
	return r.Exists(), "", nil
}

func IsOneOf(got gjson.Result, expected ...gjson.Result) (bool, string, error) {
	if len(got.Array()) == 1 {
		var message string
		ok := compareResults(got, expected)
		if !ok {
			message = "returned value `%s` not in expected values `%v`"
		}
		return ok, message, nil
	}
	results := make([]bool, len(got.Array()))
	for i, qr := range got.Array() {
		results[i] = compareResults(qr, expected)
	}
	if !allTrue(results...) {
		return false, fmt.Sprintf("returned value `%s` not in expected values `%v`", got, expected), nil
	}
	return true, "", nil
}

// func IsOneOfArray(r gjson.Result, expected ...gjson.Result) ComparisonResult {
// 	qr := r.Value()
// 	for _, exp := range expected {
// 		expectedAny, err := expectedResultsToAny(exp)
// 		if err != nil {
// 			return newComparisonResult(false, "", fmt.Errorf("could not convert expected values to any: %s", err))
// 		}
// 		ok, err := compareResults(qr.Raw, expectedAny)
// 		if err != nil {
// 			return newComparisonResult(ok, "", fmt.Errorf("could not validate query result: %s", err))
// 		}
// 		if ok {
// 			return newComparisonResult(ok, "", nil)
// 		}
// 	}
// 	return newComparisonResult(false, fmt.Sprintf("returned value `%s` not in expected values `%v`", qr, expected), nil)
// }

func compareResults(got gjson.Result, want []gjson.Result) bool {
	for _, w := range want {
		ok := reflect.DeepEqual(got.Value(), w.Value())
		if ok {
			return true
		}
	}
	return false
}

func expectedResultsToAny(in []string) ([]any, error) {
	expectedResultsAny := make([]any, 0, len(in))
	for _, exp := range in {
		var expAny any
		err := json.Unmarshal([]byte(exp), &expAny)
		if err != nil {
			var syntaxError *json.SyntaxError
			if !errors.As(err, &syntaxError) {
				return nil, fmt.Errorf("could not unmarshal expected value: %s", err)
			}
			exp2 := strconv.Quote(exp)
			err = json.Unmarshal([]byte(exp2), &expAny)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal expected value: %s", err)
			}
		}
		expectedResultsAny = append(expectedResultsAny, expAny)
	}
	return expectedResultsAny, nil
}

func allTrue(in ...bool) bool {
	for _, b := range in {
		if !b {
			return false
		}
	}
	return true
}
