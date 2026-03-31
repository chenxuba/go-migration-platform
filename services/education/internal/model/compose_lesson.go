package model

import "time"

// ComposeLessonListItem 对标 GetPageComposeLessonListForPc 列表项
type ComposeLessonListItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreateTime   time.Time `json:"createTime"`
	ProductCount int       `json:"productCount"`
	ClassCount   int       `json:"classCount"`
}
