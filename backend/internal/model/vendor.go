package model

import (
	"time"

	"gorm.io/gorm"
)

type Vendor struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Name              string         `gorm:"size:150;not null" json:"name"`
	ShortName         string         `gorm:"size:100" json:"shortName"`
	Logo              string         `gorm:"size:255" json:"logo"`
	CoverImage        string         `gorm:"size:255" json:"coverImage"`
	Province          string         `gorm:"size:50;index" json:"province"`
	City              string         `gorm:"size:50" json:"city"`
	County            string         `gorm:"size:50" json:"county"`
	Address           string         `gorm:"size:255" json:"address"`
	MainProducts      string         `gorm:"size:500" json:"mainProducts"`
	ServiceModels     string         `gorm:"size:500" json:"serviceModels"`
	ServiceAdvantages string         `gorm:"size:500" json:"serviceAdvantages"`
	Description       string         `gorm:"type:text" json:"description"`
	WebsiteURL        string         `gorm:"size:255" json:"websiteUrl"`
	Phone             string         `gorm:"size:50" json:"phone"`
	Wechat            string         `gorm:"size:100" json:"wechat"`
	ContactName       string         `gorm:"size:50" json:"contactName"`
	IsRecommended     bool           `gorm:"default:false;index" json:"isRecommended"`
	IsVerified        bool           `gorm:"default:false" json:"isVerified"`
	IsVisible         bool           `gorm:"default:true;index" json:"isVisible"`
	SortOrder         int            `gorm:"default:0" json:"sortOrder"`
	Tags              []Tag          `gorm:"many2many:vendor_tags;" json:"tags,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v Vendor) Region() string {
	if v.City == "" {
		return v.Province
	}
	return v.Province + " " + v.City
}
