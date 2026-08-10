package model

import (
	"time"

	"gorm.io/gorm"
)

type Vendor struct {
	ID                   uint             `gorm:"primaryKey" json:"id"`
	Slug                 string           `gorm:"size:180;uniqueIndex" json:"slug"`
	Name                 string           `gorm:"size:150;not null" json:"name"`
	ShortName            string           `gorm:"size:100" json:"shortName"`
	Logo                 string           `gorm:"size:255" json:"logo"`
	LogoAssetID          *uint            `gorm:"index" json:"logoAssetId,omitempty"`
	CoverImage           string           `gorm:"size:255" json:"coverImage"`
	CoverAssetID         *uint            `gorm:"index" json:"coverAssetId,omitempty"`
	Province             string           `gorm:"size:50;index" json:"province"`
	City                 string           `gorm:"size:50" json:"city"`
	County               string           `gorm:"size:50" json:"county"`
	Address              string           `gorm:"size:255" json:"address"`
	MainProducts         string           `gorm:"size:500" json:"mainProducts"`
	ServiceModels        string           `gorm:"size:500" json:"serviceModels"`
	ServiceAdvantages    string           `gorm:"size:500" json:"serviceAdvantages"`
	Description          string           `gorm:"type:text" json:"description"`
	SEOTitle             string           `gorm:"size:180" json:"seoTitle"`
	SEODescription       string           `gorm:"size:500" json:"seoDescription"`
	SEOTitleManual       bool             `gorm:"not null;default:false" json:"seoTitleManual"`
	SEODescriptionManual bool             `gorm:"not null;default:false" json:"seoDescriptionManual"`
	EstablishedYear      string           `gorm:"size:50" json:"establishedYear"`
	FactoryArea          string           `gorm:"size:100" json:"factoryArea"`
	EmployeeCount        string           `gorm:"size:100" json:"employeeCount"`
	AnnualCapacity       string           `gorm:"size:255" json:"annualCapacity"`
	Equipment            string           `gorm:"type:text" json:"equipment"`
	Certifications       string           `gorm:"type:text" json:"certifications"`
	AfterSalesService    string           `gorm:"type:text" json:"afterSalesService"`
	ReviewStatus         string           `gorm:"size:30;default:pending;index" json:"reviewStatus"`
	ProvidesProcessing   bool             `gorm:"default:false;index" json:"providesProcessing"`
	ProcessingServices   string           `gorm:"size:500" json:"processingServices"`
	ProcessingMaterials  string           `gorm:"size:500" json:"processingMaterials"`
	ProcessingEquipment  string           `gorm:"type:text" json:"processingEquipment"`
	ProcessingCapacity   string           `gorm:"size:500" json:"processingCapacity"`
	ProcessingRegions    string           `gorm:"size:500" json:"processingRegions"`
	ProcessingNotes      string           `gorm:"type:text" json:"processingNotes"`
	WebsiteURL           string           `gorm:"size:255" json:"websiteUrl"`
	Phone                string           `gorm:"size:50" json:"phone"`
	Wechat               string           `gorm:"size:100" json:"wechat"`
	WechatQRCode         string           `gorm:"size:255" json:"wechatQrCode"`
	WechatQRCodeAssetID  *uint            `gorm:"index" json:"wechatQrCodeAssetId,omitempty"`
	ContactName          string           `gorm:"size:50" json:"contactName"`
	PhonePublic          bool             `gorm:"not null;default:false" json:"phonePublic"`
	WechatPublic         bool             `gorm:"not null;default:false" json:"wechatPublic"`
	ContactNamePublic    bool             `gorm:"not null;default:false" json:"contactNamePublic"`
	PhoneAvailable       bool             `gorm:"-" json:"phoneAvailable"`
	WechatAvailable      bool             `gorm:"-" json:"wechatAvailable"`
	ContactNameAvailable bool             `gorm:"-" json:"contactNameAvailable"`
	IsRecommended        bool             `gorm:"default:false;index" json:"isRecommended"`
	IsVerified           bool             `gorm:"default:false" json:"isVerified"`
	IsVisible            bool             `gorm:"default:true;index" json:"isVisible"`
	DataOrigin           string           `gorm:"size:30;not null;default:admin;index" json:"dataOrigin"`
	PublicationStatus    string           `gorm:"size:20;not null;default:published;index" json:"publicationStatus"`
	ContentVersion       uint             `gorm:"not null;default:1" json:"contentVersion"`
	PublishedAt          *time.Time       `gorm:"index" json:"publishedAt,omitempty"`
	SortOrder            int              `gorm:"default:0" json:"sortOrder"`
	Tags                 []Tag            `gorm:"many2many:vendor_tags;" json:"tags,omitempty"`
	VendorCategories     []VendorCategory `gorm:"many2many:vendor_category_assignments;" json:"vendorCategories,omitempty"`
	Media                []VendorMedia    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"media,omitempty"`
	TagIDs               []uint           `gorm:"-" json:"tagIds,omitempty"`
	VendorCategoryIDs    []uint           `gorm:"-" json:"vendorCategoryIds,omitempty"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt   `gorm:"index" json:"-"`
}

type VendorOption struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	ShortName         string `json:"shortName"`
	Province          string `json:"province"`
	City              string `json:"city"`
	MainProducts      string `json:"mainProducts"`
	PublicationStatus string `json:"publicationStatus"`
	IsVisible         bool   `json:"isVisible"`
}

func (v Vendor) Region() string {
	if v.City == "" {
		return v.Province
	}
	return v.Province + " " + v.City
}
