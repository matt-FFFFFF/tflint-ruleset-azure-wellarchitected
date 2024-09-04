package rules

import (
	"fmt"
	"strings"

	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/blockqueryrule"
	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/ctyquery"
	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/modulecontent"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AzApiRule runs the specified gjson query on the `body` attribute of `azapi_resource` resources and checks if the result is as expected.
type AzApiRule struct {
	tflint.DefaultRule // Embed the default rule to reuse its implementation
	blockqueryrule.BlockQueryRule

	maximumApiVersion string
	minimumApiVersion string

	resourceType string
	ruleName     string
}

var _ tflint.Rule = &AzApiRule{}
var _ modulecontent.BlockFetcher = &AzApiRule{}

// AzApiRule returns a new rule.
func NewAzApiRule(ruleName, link, resourceType, minimumApiVersion, maximumApiVersion, query string, mustExist, queryResultIsArray bool, expectedResults []string) *AzApiRule {
	return &AzApiRule{
		BlockQueryRule: blockqueryrule.NewBlockQueryRule(
			ruleName,
			link,
			"resource",
			"azapi_resource",
			[]string{"type", "name"},
			"body",
		),
		expected:           expectedResults,
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

func (r *AzApiRule) LabelOne() string {
	return "azapi_resource"
}

func (r *AzApiRule) LabelNames() []string {
	return []string{"type", "name"}
}

func (r *AzApiRule) BlockType() string {
	return "resource"
}

func (r *AzApiRule) Attributes() []string {
	return []string{"name", "type", "body"}
}

func (r *AzApiRule) Check(runner tflint.Runner) error {
	return r.queryResource(runner, cty.DynamicPseudoType)
}

func (r *AzApiRule) queryResource(runner tflint.Runner, ct cty.Type) error {
	ctx, resources, diags := modulecontent.FetchBlocks(r, runner)
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
		ok, msg, err := ctyquery.Query(val, ct, r.query, r.expectedResults, r.queryResultIsArray, r.mustExist)
		if err != nil {
			return fmt.Errorf("could not query value: %s", err)
		}
		if !ok {
			runner.EmitIssue(
				r,
				msg,
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
