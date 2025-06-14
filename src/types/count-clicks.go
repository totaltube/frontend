package types

type CountClickParams struct {
	Ip        string `json:"ip"`
	Id        int64  `json:"id"`
	GroupId   int64  `json:"group_id"`
	CellIndex int64  `json:"cell_index"`
}

type CountType int

const (
	CountTypeNone CountType = iota
	CountTypeTopContent
	CountTypeTopCategories
	CountTypeCategory
	CountTypeCategoryView
)

func (c CountType) String() string {
	switch c {
	case CountTypeCategory:
		return "c"
	case CountTypeTopCategories:
		return "tca"
	case CountTypeTopContent:
		return "tc"
	case CountTypeCategoryView:
		return "ccv"
	default:
		return ""
	}
}
