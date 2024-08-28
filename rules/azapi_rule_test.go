package rules

import (
	"os"
	"testing"

	"github.com/matt-FFFFFF/tflint-ruleset-azure-wellarchitectred/modulecontent"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func TestAzapiRule(t *testing.T) {
	testCases := []struct {
		name     string
		rule     tflint.Rule
		content  string
		expected helper.Issues
	}{
		{
			name: "correct string",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"fiz"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = "fiz"
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "correct string with multiple values",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"fuz", "fiz"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = "fiz"
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "incorrect string",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"not_fiz"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = "fiz"
				bar = "biz"
			}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"not_fiz"}),
					Message: "The query `foo` returned value `fiz` not in expected values `[not_fiz]`",
				},
			},
		},
		{
			name: "correct number",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"2"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = 2
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "incorrect number",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"0"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = 2
				bar = "biz"
			}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"0"}),
					Message: "The query `foo` returned value `2` not in expected values `[0]`",
				},
			},
		},
		{
			name: "correct bool",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"true"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = true
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "incorrect bool",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"true"}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = false
				bar = "biz"
			}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, false, []string{"true"}),
					Message: "The query `foo` returned value `false` not in expected values `[true]`",
				},
			},
		},
		{
			name: "correct list",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, true, []string{`[1,2,3]`}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = [1, 2, 3]
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "incorrect list",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, true, []string{`[4,5,6]`}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = [1, 2, 3]
				bar = "biz"
			}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "foo", false, true, []string{`[4,5,6]`}),
					Message: "The query `foo` returned value `[1,2,3]` not in expected values `[[4,5,6]]`",
				},
			},
		},
		{
			name: "nested list",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo.#.bar", true, false, []string{`[4,5,6]`, `[1,2,3]`}),
			content: `
resource "azapi_resource" "test" {
	type = "testType@0000-00-00"
	body = {
			foo = [
				{
					bar = [1, 2, 3]
				},
				{
					bar = [1, 2, 3]
				},
				{
					bar = [1, 2, 3]
				}
			]
	}
}`,
			expected: helper.Issues{},
		},
		{
			name: "nested list incorrect",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "foo.#.bar", true, false, []string{`[1,2,3]`}),
			content: `
resource "azapi_resource" "test" {
type = "testType@0000-00-00"
body = {
		foo = [
			{
				bar = [1, 2, 3]
			},
			{
				bar = [4, 5, 6]
			},
			{
				bar = [1, 2, 3]
			}
		]
}
}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "foo.#.bar", true, false, []string{`[1,2,3]`}),
					Message: "The query `foo.#.bar` returned value `[[1,2,3],[4,5,6],[1,2,3]]` not in expected values `[[1,2,3]]`",
				},
			},
		},
		{
			name: "query return no results but does not need to exist",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "notexist", false, false, []string{`"fiz"`}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = "fiz"
				bar = "biz"
			}
		}`,
			expected: helper.Issues{},
		},
		{
			name: "query return no results and need to exist",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "notexist", true, false, []string{`"fiz"`}),
			content: `
		resource "azapi_resource" "test" {
		  type = "testType@0000-00-00"
		  body = {
			  foo = "fiz"
				bar = "biz"
			}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "notexist", true, false, []string{`"fiz"`}),
					Message: "The query `notexist` returned no data and `mustExist` is set",
				},
			},
		},
		{
			name: "no azapi_resources - no error expected",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "query", true, false, []string{`""`}),
			content: `
		resource "not_azapi_resource" "test" {
			biz = "baz"
			buz = "fuz"
		}`,
			expected: helper.Issues{},
		},
		{
			name: "no type attribute",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "query", true, false, []string{`""`}),
			content: `
		resource "azapi_resource" "test" {
			not_type = "baz"
			body     = {}
		}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "query", true, false, []string{`""`}),
					Message: "Resource does not have a `type` attribute",
				},
			},
		},
		{
			name: "no body attribute",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "query", true, false, []string{`""`}),
			content: `
resource "azapi_resource" "test" {
	type     = "testType@0000-00-00"
	not_body = {}
}`,
			expected: helper.Issues{
				{
					Rule:    NewAzApiRule("test", "https://example.com", "testType", "", "", "query", true, false, []string{`""`}),
					Message: "Resource does not have a `body` attribute",
				},
			},
		},
		{
			name: "object array query",
			rule: NewAzApiRule("test", "https://example.com", "testType", "", "", "objectarray.#.attr", true, false, []string{"val"}),
			content: `
resource "azapi_resource" "test" {
	type     = "testType@0000-00-00"
	body = {
		objectarray = [
		  {
				attr = "val"
			},
		  {
				attr = "val"
			}
		]
	}
}`,
			expected: helper.Issues{},
		},
	}

	filename := "main.tf"
	for _, c := range testCases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{filename: tc.content})
			stub := gostub.Stub(&modulecontent.AppFs, mockFs(tc.content))
			defer stub.Reset()
			if err := tc.rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			helper.AssertIssuesWithoutRange(t, tc.expected, runner.Issues)
		})
	}
}

func mockFs(c string) afero.Afero {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "main.tf", []byte(c), os.ModePerm)
	return afero.Afero{Fs: fs}
}
