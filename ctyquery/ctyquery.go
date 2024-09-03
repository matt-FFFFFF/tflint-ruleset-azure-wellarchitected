package ctyquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func Query(val cty.Value, ty cty.Type, query string, expected []string, expectedIsArray, mustExist bool) (ok bool, message string, err error) {
	jsonbytes, err := ctyjson.Marshal(val, ty)
	if err != nil {
		return ok, message, fmt.Errorf("could not marshal cty value: %s", err)
	}
	queryResult := gjson.GetBytes(jsonbytes, "value."+query)
	if queryResult.Exists() && len(expected) == 0 {
		message = fmt.Sprintf("The query `%s` returned data but no expected results are set", query)
		return ok, message, nil
	}
	if !queryResult.Exists() {
		if mustExist {
			message = fmt.Sprintf("The query `%s` returned no data and `mustExist` is set", query)
			return false, message, nil
		}
		return true, message, nil
	}

	expectedResultsAny, err := expectedResultsToAny(expected)
	if err != nil {
		return false, message, fmt.Errorf("could not convert expected results to any: %s", err)
	}
	if expectedIsArray {
		ok, err = validateResult(queryResult.Raw, expectedResultsAny)
		if err != nil {
			return ok, message, fmt.Errorf("could not validate query result: %s", err)
		}
		if ok {
			return ok, message, nil
		}
	} else {
		if len(queryResult.Array()) == 1 {
			ok, err = validateResult(queryResult.Raw, expectedResultsAny)
			if err != nil {
				return ok, message, fmt.Errorf("could not validate query result: %s", err)
			}
			if ok {
				return ok, message, nil
			}
		}
		results := make([]bool, len(queryResult.Array()))
		for i, qr := range queryResult.Array() {
			results[i], err = validateResult(qr.Raw, expectedResultsAny)
			if err != nil {
				return ok, message, fmt.Errorf("could not validate query result: %s", err)
			}
		}
		if allTrue(results...) {
			ok = true
		}
	}
	if !ok {
		message = fmt.Sprintf("The query `%s` returned value `%s` not in expected values `%v`", query, queryResult, expected)
	}
	return ok, message, nil
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
