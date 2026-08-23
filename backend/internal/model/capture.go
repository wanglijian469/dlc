package model

import "time"

const (
	CaptureStatusUploading   = "uploading"
	CaptureStatusQueued      = "queued"
	CaptureStatusProcessing  = "processing"
	CaptureStatusNeedsReview = "needs_review"
	CaptureStatusReady       = "ready"
	CaptureStatusCommitted   = "committed"
	CaptureStatusFailed      = "failed"
)

// CapturePackage groups every photo from one vendor. DraftJSON is deliberately
// independent from published records so recognition can be rerun safely.
type CapturePackage struct {
	ID                 uint                 `gorm:"primaryKey" json:"id"`
	Title              string               `gorm:"size:150;not null" json:"title"`
	OwnerUsername      string               `gorm:"size:100;not null;index" json:"ownerUsername"`
	OwnerRole          string               `gorm:"size:20;not null;index" json:"ownerRole"`
	VendorID           *uint                `gorm:"index" json:"vendorId,omitempty"`
	Vendor             *Vendor              `json:"vendor,omitempty"`
	Status             string               `gorm:"size:30;not null;default:uploading;index" json:"status"`
	RecognitionVersion uint                 `gorm:"not null;default:0" json:"recognitionVersion"`
	DraftJSON          string               `gorm:"type:longtext" json:"-"`
	ErrorMessage       string               `gorm:"type:text" json:"errorMessage,omitempty"`
	Attempts           int                  `gorm:"not null;default:0" json:"attempts"`
	QueuedAt           *time.Time           `json:"queuedAt,omitempty"`
	StartedAt          *time.Time           `json:"startedAt,omitempty"`
	CompletedAt        *time.Time           `json:"completedAt,omitempty"`
	CommittedAt        *time.Time           `json:"committedAt,omitempty"`
	Documents          []CaptureDocument    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"documents,omitempty"`
	Crops              []CaptureProductCrop `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"crops,omitempty"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
}

type CaptureDocument struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CapturePackageID  uint       `gorm:"not null;index" json:"capturePackageId"`
	AssetID           uint       `gorm:"not null;uniqueIndex" json:"assetId"`
	Asset             MediaAsset `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"asset"`
	DocumentType      string     `gorm:"size:30;not null;default:unknown;index" json:"documentType"`
	SortOrder         int        `gorm:"not null;default:0" json:"sortOrder"`
	OCRJSON           string     `gorm:"type:longtext" json:"-"`
	OCRText           string     `gorm:"type:longtext" json:"ocrText,omitempty"`
	OCRConfidence     float64    `gorm:"not null;default:0" json:"ocrConfidence"`
	ProviderRequestID string     `gorm:"size:100" json:"providerRequestId,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type CaptureResult struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	CapturePackageID uint      `gorm:"not null;index" json:"capturePackageId"`
	Version          uint      `gorm:"not null" json:"version"`
	Provider         string    `gorm:"size:30;not null" json:"provider"`
	Model            string    `gorm:"size:100" json:"model"`
	ResultJSON       string    `gorm:"type:longtext;not null" json:"-"`
	CreatedAt        time.Time `json:"createdAt"`
}

type CaptureProductCrop struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	CapturePackageID uint       `gorm:"not null;index" json:"capturePackageId"`
	DocumentID       uint       `gorm:"not null;index" json:"documentId"`
	AssetID          uint       `gorm:"not null;index" json:"assetId"`
	Asset            MediaAsset `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"asset"`
	ProductKey       string     `gorm:"size:100;index" json:"productKey"`
	X                float64    `gorm:"not null" json:"x"`
	Y                float64    `gorm:"not null" json:"y"`
	Width            float64    `gorm:"not null" json:"width"`
	Height           float64    `gorm:"not null" json:"height"`
	Selected         bool       `gorm:"not null;default:false" json:"selected"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type VendorInvitation struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	VendorID  uint       `gorm:"not null;index" json:"vendorId"`
	Vendor    Vendor     `json:"vendor"`
	TokenHash string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	CreatedBy string     `gorm:"size:100;not null" json:"createdBy"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}
