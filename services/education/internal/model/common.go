package model

type PageRequestModel struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type SortModel struct {
	ByUpdateTime        int `json:"byUpdateTime"`
	ByTotalSales        int `json:"byTotalSales"`
	OrderBySortNo       int `json:"orderBySortNumber"`
	ByCreatedTime       int `json:"byCreatedTime"`
	ByFollowUpTime      int `json:"byFollowUpTime"`
	ByNextFlowTime      int `json:"byNextFlowTime"`
	ByDaysUntilReturn   int `json:"byDaysUntilReturn"`
	BySalesAssignedTime int `json:"bySalesAssignedTime"`
}

type PageResult[T any] struct {
	Items   []T `json:"items"`
	Total   int `json:"total"`
	Current int `json:"current"`
	Size    int `json:"size"`
}
