package blockqueryrule

import (
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type BlockQueryRule struct {
	tflint.DefaultRule
	name            string
	blockType       string
	labelOne        string
	blockLabelNames []string
	query           string
	link            string
	queryAttribute  string
}

func NewBlockQueryRule(ruleName, link, blockType, labelOne string, blockLabelNames []string, queryAttribute, query string, queryResultIsArray, queryRestMustExist bool, expected []string) BlockQueryRule {
	return BlockQueryRule{
		name:            ruleName,
		blockType:       blockType,
		blockLabelNames: blockLabelNames,
		labelOne:        labelOne,
		queryAttribute:  queryAttribute,
		link:            link,
		query:           query,
	}
}
