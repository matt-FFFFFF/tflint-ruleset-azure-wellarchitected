package azapi

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

func loadResources(typeAttr string, runner tflint.Runner) (*terraform.Evaluator, []*hclext.Block, hcl.Diagnostics) {
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

	resources, diags := filterResourceBlocksByType(ctx, config.Module, typeAttr)

	return ctx, resources, diags
}

func filterResourceBlocksByType(ctx *terraform.Evaluator, module *terraform.Module, resourceType string) ([]*hclext.Block, hcl.Diagnostics) {
	resources, diags := getResourceBlocks(ctx, module)
	if diags.HasErrors() {
		return nil, diags
	}
	filteredResources := make([]*hclext.Block, 0, len(resources.Blocks))
	for _, resource := range resources.Blocks {

		if resource.Labels[0] != "azapi_resource" {
			continue
		}
		filteredResources = append(filteredResources, resource)
	}
	return filteredResources, nil
}

func getResourceBlocks(ctx *terraform.Evaluator, module *terraform.Module) (*hclext.BodyContent, hcl.Diagnostics) {
	resources, diags := module.PartialContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{
							Name:     "type",
							Required: false,
						},
						{
							Name:     "body",
							Required: false,
						},
					},
				},
			},
		},
	}, ctx)

	return resources, diags
}
