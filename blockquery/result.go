package blockquery

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

func NewResult(ty gjson.Type, val any) gjson.Result {
	raw, err := json.Marshal(val)
	if err != nil {
		panic(fmt.Sprintf("could not marshal value: %s", err))
	}
	res := gjson.Result{
		Raw:  string(raw),
		Type: ty,
	}
	switch ty {
	case gjson.Number:
		res.Num = val.(float64)
	case gjson.String:
		res.Str = val.(string)
	case gjson.JSON:
		res.Raw = string(raw)
	}
	return res
}

func NewIntResults(vals ...int) []gjson.Result {
	results := make([]gjson.Result, len(vals))
	for i, val := range vals {
		results[i] = NewResult(gjson.Number, val)
	}
	return results
}

func NewFloatResults(vals ...float64) []gjson.Result {
	results := make([]gjson.Result, len(vals))
	for i, val := range vals {
		results[i] = NewResult(gjson.Number, val)
	}
	return results
}

func NewStringResults(vals ...string) []gjson.Result {
	results := make([]gjson.Result, len(vals))
	for i, val := range vals {
		results[i] = NewResult(gjson.String, val)
	}
	return results
}

func NewTrueResults(vals ...bool) []gjson.Result {
	return []gjson.Result{NewResult(gjson.True, true)}
}

func NewFalseResults(vals ...bool) []gjson.Result {
	return []gjson.Result{NewResult(gjson.False, false)}
}

func NewJsonResults(vals ...any) []gjson.Result {
	results := make([]gjson.Result, len(vals))
	for i, val := range vals {
		results[i] = NewResult(gjson.JSON, val)
	}
	return results
}
