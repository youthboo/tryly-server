package dto

// Certificate DTOs
type CreateCertificateRequest struct {
	CertificateID  int64  `json:"certificate_id" validate:"gt=0"`
	CertificateName string `json:"certificate_name" validate:"notblank"`
	ImageURL       string `json:"image_url"`
	ExpiryDate     *string `json:"expiry_date"` // YYYY-MM-DD
	IssueDate      *string `json:"issue_date"`  // YYYY-MM-DD
	Notes          *string `json:"notes"`
}

type PatchCertificateRequest struct {
	DocumentURL *string `json:"document_url"`
	ExpireDate  *string `json:"expire_date"`
	CertNumber  *string `json:"cert_number"`
}

// Message DTOs
type CreateMessageRequest struct {
	ReferenceType *string `json:"reference_type"`
	ReferenceID   *int64  `json:"reference_id"`
	ReceiverID    *int64  `json:"receiver_id"`
	Content       *string `json:"content"`
	AttachmentURL *string `json:"attachment_url"`
	ConvID        *int64  `json:"conv_id"`
	MessageType   *string `json:"message_type"`
	QuoteData     *string `json:"quote_data"`
}

// Conversation DTOs
type CreateConversationRequest struct {
	ParticipantID int64  `json:"participant_id" validate:"gt=0"`
	Message       string `json:"message" validate:"notblank"`
	ReferenceType *string `json:"reference_type"`
	ReferenceID   *int64  `json:"reference_id"`
}

type ShareRFQRequest struct {
	RFQID   int64  `json:"rfq_id" validate:"gt=0"`
	Message *string `json:"message"`
}

type MarkConversationAsReadRequest struct {
	IDs []int64 `json:"ids"`
}

// Production DTOs
type CreateProductionUpdateRequest struct {
	StepID          int64    `json:"step_id" validate:"gte=0"` // gte=0: step_id=0 คือขั้น "ยืนยันรับงาน" พิเศษ
	Status          string   `json:"status" validate:"notblank"`
	CompletedAt     *string  `json:"completed_at"` // RFC3339
	Notes           *string  `json:"notes"`
	ImageURLs       []string `json:"image_urls"`
	ProgressPercent *int     `json:"progress_percent"`
	// step_id=4 (จัดส่งแล้ว): บันทึก tracking_no / courier ลง orders table
	TrackingNo *string `json:"tracking_no"`
	Courier    *string `json:"courier"`
}

type RejectProductionUpdateRequest struct {
	Reason string `json:"reason" validate:"notblank"`
	Notes  *string `json:"notes"`
}

type CreateProductionStepRequest struct {
	OrderID     int64  `json:"order_id" validate:"gt=0"`
	StepName    string `json:"step_name" validate:"notblank"`
	Description *string `json:"description"`
	SequenceNo  int    `json:"sequence_no"`
	EstimatedDays *int `json:"estimated_days"`
}
