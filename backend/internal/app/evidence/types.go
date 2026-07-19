package evidence

import "errors"

var (
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidState = errors.New("invalid state")
	ErrNotFound     = errors.New("not found")
)

const (
	maxAttachmentsPerPayment = 3
	maxAttachmentSize        = int64(5 << 20)
)

type Config struct {
	DataRoot string
	Storage  *Storage
}

type Attachment struct {
	ID         int64  `json:"id"`
	PaymentID  int64  `json:"paymentId"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	FileSize   int64  `json:"fileSize"`
	UploadedBy int64  `json:"uploadedBy"`
	UploadedAt string `json:"uploadedAt"`
	FilePath   string `json:"-"`
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}
