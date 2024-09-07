package blockquery

type BlockQuery struct {
	BlockType       string
	LabelOne        string
	BlockLabelNames []string
	Query           string
	QueryAttribute  string
	CompareFunc     ResultCompareFunc
}

func NewBlockQueryRule(
	blockType, labelOne string,
	blockLabelNames []string,
	queryAttribute, query string,
	cmpFn ResultCompareFunc) BlockQuery {
	return BlockQuery{
		BlockType:       blockType,
		BlockLabelNames: blockLabelNames,
		LabelOne:        labelOne,
		QueryAttribute:  queryAttribute,
		Query:           query,
		CompareFunc:     cmpFn,
	}
}
