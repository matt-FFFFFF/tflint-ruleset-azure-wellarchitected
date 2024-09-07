package blockquery

type BlockQuery struct {
	blockType       string
	labelOne        string
	blockLabelNames []string
	query           string
	queryAttribute  string
	compareFunc     ResultCompareFunc
}

func NewBlockQueryRule(
	blockType, labelOne string,
	blockLabelNames []string,
	queryAttribute, query string,
	cmpFn ResultCompareFunc) BlockQuery {
	return BlockQuery{
		blockType:       blockType,
		blockLabelNames: blockLabelNames,
		labelOne:        labelOne,
		queryAttribute:  queryAttribute,
		query:           query,
		compareFunc:     cmpFn,
	}
}
