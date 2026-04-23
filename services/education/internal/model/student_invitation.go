package model

type PendingAttentionInvitationQRCodeDTO struct {
	StudentID int64 `json:"studentId"`
}

type PendingAttentionInvitationQRCodeVO struct {
	StudentID           int64  `json:"studentId"`
	StudentName         string `json:"studentName"`
	InstitutionID       int64  `json:"institutionId"`
	InstitutionName     string `json:"institutionName"`
	OfficialAccountName string `json:"officialAccountName"`
	SceneValue          string `json:"sceneValue"`
	QRCodeURL           string `json:"qrCodeUrl"`
	QRCodeDataURL       string `json:"qrCodeDataUrl,omitempty"`
}
