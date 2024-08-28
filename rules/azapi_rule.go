package rules

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/modulecontent"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/tidwall/gjson"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// AzapiRule runs the specified gjson query on the `body` attribute of `azapi_resource` resources and checks if the result is as expected.
type AzapiRule struct {
	tflint.DefaultRule // Embed the default rule to reuse its implementation
	apiVersion         string
	expectedResults    []string
	link               string
	mustExist          bool
	query              string
	resourceType       string
	ruleName           string
}

var _ tflint.Rule = &AzapiRule{}
var _ modulecontent.ContentFetcher = &AzapiRule{}

// NewNewAzapiRule returns a new rule with the given resource type, attribute name, and expected values.
func NewAzapiRule(ruleName, link, resourceType, apiVersion, query string, mustExist bool, expectedResults []string) *AzapiRule {
	return &AzapiRule{
		apiVersion:      apiVersion,
		expectedResults: expectedResults,
		link:            link,
		mustExist:       mustExist,
		query:           query,
		resourceType:    resourceType,
		ruleName:        ruleName,
	}
}

func (r *AzapiRule) Link() string {
	return r.link
}

func (r *AzapiRule) Enabled() bool {
	return true
}

func (r *AzapiRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *AzapiRule) Name() string {
	return r.ruleName
}

func (r *AzapiRule) ResourceType() string {
	return "azapi_resource"
}

func (r *AzapiRule) Attributes() []string {
	return []string{"name", "type", "body"}
}

func (r *AzapiRule) Check(runner tflint.Runner) error {
	return r.queryResource(runner, cty.DynamicPseudoType)
}

func (r *AzapiRule) queryResource(runner tflint.Runner, ct cty.Type) error {
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
		if !checkAzApiType(typeStr, r.resourceType, r.apiVersion) {
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
		var expectedIsArray bool
		var expectedResultsAny []any
		for _, exp := range r.expectedResults {
			var expAny any
			err := json.Unmarshal([]byte(exp), &expAny)
			if err != nil {
				return fmt.Errorf("could not unmarshal expected value: %s", err)
			}
			if _, ok := expAny.([]any); ok {
				expAny = expAny.([]any)
				expectedIsArray = true
			}
			expectedResultsAny = append(expectedResultsAny, expAny)
		}
		if expectedIsArray {
			for _, exp := range expectedResultsAny {
				var queryResultAny []any
				err := json.Unmarshal([]byte(queryResult.Raw), &queryResultAny)
				if err != nil {
					return fmt.Errorf("could not unmarshal query result: %s", err)
				}
				if reflect.DeepEqual(exp, queryResultAny) {
					ok = true
					break
				}
			}
		} else {
			for _, qr := range queryResult.Array() {
				for _, exp := range expectedResultsAny {
					var queryResultAny []any
					err := json.Unmarshal([]byte(qr.Raw), &queryResultAny)
					if err != nil {
						return fmt.Errorf("could not unmarshal query result: %s", err)
					}
					if reflect.DeepEqual(exp, queryResultAny) {
						ok = true
						break
					}
				}
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

func checkAzApiType(gotType, wantResourceType, wantApiVersion string) bool {
	gotSplit := strings.Split(gotType, "@")
	if !strings.EqualFold(gotSplit[0], wantResourceType) {
		return false
	}
	if wantApiVersion == "" || strings.EqualFold(gotSplit[1], wantApiVersion) {
		return true
	}
	return false
}
