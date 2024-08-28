package modulecontent

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/terraform/addrs"
)

var AppFs = afero.Afero{
	Fs: afero.NewOsFs(),
}

type ContentFetcher interface {
	ResourceType() string
	Attributes() []string
}

func FetchAttributes(f ContentFetcher, runner tflint.Runner) (*terraform.Evaluator, []*hclext.Attribute, hcl.Diagnostics) {
	config, ctx, diags := initEvaluator(runner)
	if diags.HasErrors() {
		return nil, nil, diags
	}
	attrs, diags := getAttributes(config.Module, f.ResourceType(), f.Attributes(), ctx)
	return ctx, attrs, diags
}

func FetchResources(f ContentFetcher, runner tflint.Runner) (*terraform.Evaluator, []*hclext.Block, hcl.Diagnostics) {
	config, ctx, diags := initEvaluator(runner)
	if diags.HasErrors() {
		return nil, nil, diags
	}
	resources, diags := getResources(ctx, config.Module, f.ResourceType(), f.Attributes())

	return ctx, resources, diags
}

// getSimpleResources returns a slice of resources with the given resource type and the attribute if it exists.
func getResources(ctx *terraform.Evaluator, module *terraform.Module, resourceType string, attributes []string) ([]*hclext.Block, hcl.Diagnostics) {
	resources, diags := getResourcesOfResourceTypeIncludingSpecifiedAttribute(module, attributes, ctx)
	if diags.HasErrors() {
		return nil, diags
	}
	filteredResources := make([]*hclext.Block, 0, len(resources.Blocks))
	for _, resource := range resources.Blocks {
		if resource.Labels[0] != resourceType {
			continue
		}
		filteredResources = append(filteredResources, resource)
	}
	return filteredResources, nil
}

// getAttributes returns a slice of attributes with the given attribute name from the resources of the given resource type.
func getAttributes(module *terraform.Module, resourceType string, attributes []string, ctx *terraform.Evaluator) ([]*hclext.Attribute, hcl.Diagnostics) {
	resources, diags := getResourcesOfResourceTypeIncludingSpecifiedAttribute(module, attributes, ctx)
	if diags.HasErrors() {
		return nil, diags
	}
	attrs := make([]*hclext.Attribute, 0, len(resources.Blocks))
	for _, resource := range resources.Blocks {
		if resource.Labels[0] != resourceType {
			continue
		}
		for _, attribute := range attributes {
			if attribute := getAttrFromBlock(resource, attribute); attribute != nil {
				attrs = append(attrs, attribute)
			}
		}
	}
	return attrs, nil
}

// getAttrFromBlock returns the attribute with the given attribute name from the block.
func getAttrFromBlock(block *hclext.Block, attributeName string) *hclext.Attribute {
	attribute, exists := block.Body.Attributes[attributeName]
	if !exists {
		return nil
	}
	return attribute
}

func initEvaluator(runner tflint.Runner) (*terraform.Config, *terraform.Evaluator, hcl.Diagnostics) {
	wd, _ := runner.GetOriginalwd()
	loader, err := terraform.NewLoader(AppFs, wd)
	if err != nil {
		return nil, nil, hcl.Diagnostics{{
			Summary: err.Error(),
		}}
	}
	config, diags := loader.LoadConfig(".", terraform.CallLocalModule)
	if diags.HasErrors() {
		return nil, nil, diags
	}
	vvals, diags := terraform.VariableValues(config)
	if diags.HasErrors() {
		return nil, nil, diags
	}
	ctx := &terraform.Evaluator{
		Meta: &terraform.ContextMeta{
			Env:                "",
			OriginalWorkingDir: wd,
		},
		Config:         config,
		VariableValues: vvals,
		ModulePath:     addrs.RootModuleInstance,
	}
	return config, ctx, nil
}

func getResourcesOfResourceTypeIncludingSpecifiedAttribute(module *terraform.Module, attributes []string, ctx *terraform.Evaluator) (*hclext.BodyContent, hcl.Diagnostics) {
	attrSchema := make([]hclext.AttributeSchema, 0, len(attributes))
	for _, attr := range attributes {
		attrSchema = append(attrSchema, hclext.AttributeSchema{
			Name:     attr,
			Required: false,
		})
	}
	resources, diags := module.PartialContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: attrSchema,
				},
			},
		},
	}, ctx)
	return resources, diags
}
