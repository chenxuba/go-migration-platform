package model

import "time"

type FaceCollectionStudentQueryDTO struct {
	PageRequestModel PageRequestModel            `json:"pageRequestModel"`
	QueryModel       FaceCollectionStudentFilter `json:"queryModel"`
}

type FaceCollectionStudentFilter struct {
	SearchKey string `json:"searchKey"`
}

type FaceCollectionStudent struct {
	ID        int64  `json:"id"`
	StuName   string `json:"stuName"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Mobile    string `json:"mobile"`
	IsCollect bool   `json:"isCollect"`
}

type FaceCollectionProfile struct {
	StudentID      int64      `json:"studentId"`
	StuName        string     `json:"stuName,omitempty"`
	FaceDescriptor []float32  `json:"faceDescriptor,omitempty"`
	FaceImage      string     `json:"faceImage,omitempty"`
	UpdatedTime    *time.Time `json:"updatedTime,omitempty"`
}

type FaceCollectionProfileSaveDTO struct {
	StudentID      int64     `json:"studentId"`
	FaceDescriptor []float32 `json:"faceDescriptor"`
	FaceImage      string    `json:"faceImage"`
}

type FaceCollectionProfileDeleteDTO struct {
	StudentID int64 `json:"studentId"`
}

type FaceAttendanceRecord struct {
	ID          int64      `json:"id"`
	StudentID   int64      `json:"studentId"`
	StudentName string     `json:"studentName"`
	FaceImage   string     `json:"faceImage,omitempty"`
	RecordTime  *time.Time `json:"recordTime,omitempty"`
}

type FaceAttendanceRecordSaveDTO struct {
	StudentID int64  `json:"studentId"`
	FaceImage string `json:"faceImage"`
}
