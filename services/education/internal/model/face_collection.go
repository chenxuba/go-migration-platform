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

type FaceCollectionCompareDTO struct {
	FaceDescriptor []float32 `json:"faceDescriptor"`
}

type FaceCollectionCompareResult struct {
	Matched     bool    `json:"matched"`
	StudentID   int64   `json:"studentId,omitempty"`
	StudentName string  `json:"studentName,omitempty"`
	Distance    float64 `json:"distance,omitempty"`
}

const (
	FaceAttendanceSessionStatusSignedIn  = 1
	FaceAttendanceSessionStatusSignedOut = 2

	FaceAttendanceSessionActionSignIn           = "sign_in"
	FaceAttendanceSessionActionSignOut          = "sign_out"
	FaceAttendanceSessionActionDuplicateSignIn  = "duplicate_sign_in"
	FaceAttendanceSessionActionDuplicateSignOut = "duplicate_sign_out"
	FaceAttendanceSessionActionIgnore           = "ignore"
	FaceAttendanceSessionActionNoMatch          = "no_match"
)

type FaceAttendanceSession struct {
	ID                int64      `json:"id"`
	StudentID         int64      `json:"studentId"`
	StudentName       string     `json:"studentName"`
	AvatarURL         string     `json:"avatarUrl,omitempty"`
	AttendanceDate    string     `json:"attendanceDate"`
	SessionNo         int        `json:"sessionNo"`
	Status            int        `json:"status"`
	SignInTime        *time.Time `json:"signInTime,omitempty"`
	SignInImage       string     `json:"signInImage,omitempty"`
	SignOutTime       *time.Time `json:"signOutTime,omitempty"`
	SignOutImage      string     `json:"signOutImage,omitempty"`
	LatestAction      string     `json:"latestAction"`
	LatestActionLabel string     `json:"latestActionLabel"`
	LatestTime        *time.Time `json:"latestTime,omitempty"`
	LatestImage       string     `json:"latestImage,omitempty"`
	HasSchedule       bool       `json:"hasSchedule"`
	LastLessonEndTime *time.Time `json:"lastLessonEndTime,omitempty"`
}

type FaceAttendanceSessionRecognizeDTO struct {
	FaceDescriptor []float32 `json:"faceDescriptor"`
}

type FaceAttendanceSessionRecognizeResult struct {
	Matched           bool       `json:"matched"`
	StudentID         int64      `json:"studentId,omitempty"`
	StudentName       string     `json:"studentName,omitempty"`
	AvatarURL         string     `json:"avatarUrl,omitempty"`
	Distance          float64    `json:"distance,omitempty"`
	Action            string     `json:"action"`
	ActionLabel       string     `json:"actionLabel"`
	SessionID         int64      `json:"sessionId,omitempty"`
	SessionNo         int        `json:"sessionNo,omitempty"`
	NeedUpload        bool       `json:"needUpload"`
	Message           string     `json:"message,omitempty"`
	HasSchedule       bool       `json:"hasSchedule"`
	LastActionTime    *time.Time `json:"lastActionTime,omitempty"`
	LastLessonEndTime *time.Time `json:"lastLessonEndTime,omitempty"`
}

type FaceAttendanceSessionCommitDTO struct {
	StudentID int64  `json:"studentId"`
	SessionID int64  `json:"sessionId"`
	Action    string `json:"action"`
	FaceImage string `json:"faceImage"`
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
