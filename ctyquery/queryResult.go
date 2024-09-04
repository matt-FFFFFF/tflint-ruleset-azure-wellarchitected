package ctyquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/tidwall/gjson"
)

type ResultCompareFunc func(gjson.Result, ...gjson.Result) ComparisonResult

func Exists(r gjson.Result, _ ...gjson.Result) ComparisonResult {
	return newComparisonResult(r.Exists(), "", nil)
}

func IsOneOf(result gjson.Result, expected ...gjson.Result) ComparisonResult {
	qr := m.Value()
	if len(qr.Array()) == 1 {
		result.
			ok, err := validateResult(m.Value().Raw, expectedAny)
		if err != nil {
			return newComparisonResult(ok, "", fmt.Errorf("could not validate query result: %s", err))
		}
		if ok {
			return newComparisonResult(ok, "", nil)
		}
	}
	results := make([]bool, len(qr.Array()))
	for i, qr := range qr.Array() {
		results[i], err = validateResult(qr.Raw, expectedAny)
		if err != nil {
			return newComparisonResult(false, "", fmt.Errorf("could not validate query result: %s", err))
		}
	}
	if !allTrue(results...) {
		return newComparisonResult(false, fmt.Sprintf("returned value `%s` not in expected values `%v`", qr, expected), nil)
	}
	return newComparisonResult(true, "", nil)
}

func (m QueryResult) IsOneOfArray(expected [][]string) ComparisonResult {
	qr := m.Value()
	for _, exp := range expected {
		expectedAny, err := expectedResultsToAny(exp)
		if err != nil {
			return newComparisonResult(false, "", fmt.Errorf("could not convert expected values to any: %s", err))
		}
		ok, err := validateResult(qr.Raw, expectedAny)
		if err != nil {
			return newComparisonResult(ok, "", fmt.Errorf("could not validate query result: %s", err))
		}
		if ok {
			return newComparisonResult(ok, "", nil)
		}
	}
	return newComparisonResult(false, fmt.Sprintf("returned value `%s` not in expected values `%v`", qr, expected), nil)
}

func validateResult(got string, want []any) (bool, error) {
	var gotAny any
	err := json.Unmarshal([]byte(got), &gotAny)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal query result: %s", err)
	}
	for _, w := range want {
		if reflect.DeepEqual(gotAny, w) {
			return true, nil
		}
	}
	return false, nil
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
