package blockqueryrule

import (
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type BlockQueryRule struct {
	tflint.DefaultRule
	name               string
	blockType          string
	labelOne           string
	blockLabelNames    []string
	expected           []string
	query              string
	queryResultIsArray bool
	queryRestMustExist bool
	link               string
	queryAttribute     string
}

func NewBlockQueryRule(ruleName, link, blockType, labelOne string, blockLabelNames []string, queryAttribute, query string, queryResultIsArray, queryRestMustExist bool, expected []string) *BlockQueryRule {
	return &BlockQueryRule{
		name:               ruleName,
		blockType:          blockType,
		blockLabelNames:    blockLabelNames,
		labelOne:           labelOne,
		queryAttribute:     queryAttribute,
		expected:           expected,
		link:               link,
		query:              query,
		queryResultIsArray: queryResultIsArray,
		queryRestMustExist: queryRestMustExist,
	}
}
