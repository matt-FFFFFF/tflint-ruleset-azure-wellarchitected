package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/modulecontent"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/tidwall/gjson"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// AzApiRule runs the specified gjson query on the `body` attribute of `azapi_resource` resources and checks if the result is as expected.
type AzApiRule struct {
	tflint.DefaultRule // Embed the default rule to reuse its implementation
	expectedResults    []string
	link               string
	maximumApiVersion  string
	minimumApiVersion  string
	mustExist          bool
	query              string
	queryResultIsArray bool
	resourceType       string
	ruleName           string
}

var _ tflint.Rule = &AzApiRule{}
var _ modulecontent.ContentFetcher = &AzApiRule{}

// AzApiRule returns a new rule.
func NewAzApiRule(ruleName, link, resourceType, minimumApiVersion, maximumApiVersion, query string, mustExist, queryResultIsArray bool, expectedResults []string) *AzApiRule {
	return &AzApiRule{
		expectedResults:    expectedResults,
		link:               link,
		maximumApiVersion:  maximumApiVersion,
		minimumApiVersion:  minimumApiVersion,
		mustExist:          mustExist,
		query:              query,
		queryResultIsArray: queryResultIsArray,
		resourceType:       resourceType,
		ruleName:           ruleName,
	}
}

func (r *AzApiRule) Link() string {
	return r.link
}

func (r *AzApiRule) Enabled() bool {
	return true
}

func (r *AzApiRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *AzApiRule) Name() string {
	return r.ruleName
}

func (r *AzApiRule) ResourceType() string {
	return "azapi_resource"
}

func (r *AzApiRule) Attributes() []string {
	return []string{"name", "type", "body"}
}

func (r *AzApiRule) Check(runner tflint.Runner) error {
	return r.queryResource(runner, cty.DynamicPseudoType)
}

func (r *AzApiRule) queryResource(runner tflint.Runner, ct cty.Type) error {
	ctx, resources, diags := modulecontent.FetchResources(r, runner)
	if diags.HasErrors() {
		return fmt.Errorf("could not get partial content: %s", diags)
	}
	for _, resource := range resources {
		typeAttr, typeAttrExists := resource.Body.Attributes["type"]
		if !typeAttrExists {
			runner.EmitIssue(
				r,
				"Resource does not have a `type` attribute",
				resource.DefRange,
			)
			continue
		}
		typeVal, diags := ctx.EvaluateExpr(typeAttr.Expr, cty.String)
		if diags.HasErrors() {
			return fmt.Errorf("could not evaluate type expression: %s", diags)
		}
		typeStr := typeVal.AsString()
		if !checkAzApiType(typeStr, r.resourceType, r.minimumApiVersion, r.maximumApiVersion) {
			continue
		}
		bodyAttr, bodyAttrExists := resource.Body.Attributes["body"]
		if !bodyAttrExists {
			runner.EmitIssue(
				r,
				"Resource does not have a `body` attribute",
				resource.DefRange,
			)
			continue
		}

		val, diags := ctx.EvaluateExpr(bodyAttr.Expr, ct)
		if diags.HasErrors() {
			return fmt.Errorf("could not evaluate body expression: %s", diags)
		}
		jsonbytes, err := ctyjson.Marshal(val, ct)
		if err != nil {
			return fmt.Errorf("could not marshal cty value: %s", err)
		}
		queryResult := gjson.GetBytes(jsonbytes, "value."+r.query)
		if !queryResult.Exists() {
			if r.mustExist {
				runner.EmitIssue(
					r,
					fmt.Sprintf("The query `%s` returned no data and `mustExist` is set", r.query),
					bodyAttr.Range,
				)
			}
			continue
		}
		var ok bool
		expectedResultsAny, err := expectedResultsToAny(r.expectedResults)
		if err != nil {
			return fmt.Errorf("could not convert expected results to any: %s", err)
		}
		if r.queryResultIsArray {
			ok, err = validateResult(queryResult.Raw, expectedResultsAny)
			if err != nil {
				return fmt.Errorf("could not validate query result: %s", err)
			}
			if ok {
				continue
			}
		} else {
			if len(queryResult.Array()) == 1 {
				ok, err = validateResult(queryResult.Raw, expectedResultsAny)
				if err != nil {
					return fmt.Errorf("could not validate query result: %s", err)
				}
				if ok {
					continue
				}
			}
			results := make([]bool, len(queryResult.Array()))
			for i, qr := range queryResult.Array() {
				results[i], err = validateResult(qr.Raw, expectedResultsAny)
				if err != nil {
					return fmt.Errorf("could not validate query result: %s", err)
				}
			}
			if allTrue(results...) {
				ok = true
			}
		}
		if !ok {
			runner.EmitIssue(
				r,
				fmt.Sprintf("The query `%s` returned value `%s` not in expected values `%v`", r.query, queryResult, r.expectedResults),
				bodyAttr.Range,
			)
		}

	}
	return nil
}

func checkAzApiType(gotType, wantResourceType, minimumApiVersion, maximumApiVersion string) bool {
	gotSplit := strings.Split(gotType, "@")
	if len(gotSplit) != 2 {
		return false
	}
	if !strings.EqualFold(gotSplit[0], wantResourceType) {
		return false
	}
	if minimumApiVersion != "" {
		if gotSplit[1] < minimumApiVersion {
			return false
		}
	}
	if maximumApiVersion != "" {
		if gotSplit[1] > maximumApiVersion {
			return false
		}
	}
	return true
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
