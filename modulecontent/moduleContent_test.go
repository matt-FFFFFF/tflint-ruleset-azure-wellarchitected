package modulecontent

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
)

func TestFetchBlocks(t *testing.T) {
	mockRunner := new(MockRunner)
	mockBlockFetcher := new(MockBlockFetcher)

	mockBlockFetcher.On("BlockType").Return("resource")
	mockBlockFetcher.On("LabelOne").Return("azapi_resource")
	mockBlockFetcher.On("LabelNames").Return([]string{"type", "name"})
	mockBlockFetcher.On("Attributes").Return([]string{"name", "type"})

	config, ctx, diags := initEvaluator(mockRunner)
	if diags.HasErrors() {
		t.Fatalf("Failed to initialize evaluator: %v", diags)
	}

	// Mock the module content
	module := &terraform.Module{
		Blocks: []*hclext.Block{
			{
				Labels: []string{"azapi_resource", "example"},
				Body: &hclext.Body{
					Attributes: map[string]*hclext.Attribute{
						"name": {Name: "name"},
						"type": {Name: "type"},
					},
				},
			},
		},
	}

	ctx.Config.Module = module

	_, blocks, diags := FetchBlocks(mockBlockFetcher, mockRunner)
	if diags.HasErrors() {
		t.Fatalf("FetchBlocks returned errors: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Labels[0] != "azapi_resource" {
		t.Errorf("Expected block label 'azapi_resource', got '%s'", blocks[0].Labels[0])
	}
}

// MockRunner is a mock implementation of tflint.Runner for testing purposes.
type MockRunner struct {
	tflint.Runner
}

func (m *MockRunner) GetOriginalwd() (string, error) {
	return "/mock/path", nil
}

// MockBlockFetcher is a mock implementation of BlockFetcher for testing purposes.
type MockBlockFetcher struct {
	BlockFetcher
	mock.Mock
}

func (m *MockBlockFetcher) BlockType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBlockFetcher) LabelOne() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBlockFetcher) LabelNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockBlockFetcher) Attributes() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
