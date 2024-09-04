package ctyquery

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func Query(val cty.Value, ty cty.Type, query string, expected []string, expectedIsArray, mustExist bool) QueryResult {
	jsonbytes, err := ctyjson.Marshal(val, ty)
	if err != nil {
		return newQueryResult(gjson.Result{}, fmt.Errorf("could not marshal cty value: %s", err))
	}
	return newQueryResult(gjson.GetBytes(jsonbytes, "value."+query), nil)
}
