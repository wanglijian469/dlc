package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ContentBlock struct {
	Type       string   `json:"type"`
	Title      string   `json:"title,omitempty"`
	Text       string   `json:"text,omitempty"`
	Items      []string `json:"items,omitempty"`
	ButtonText string   `json:"buttonText,omitempty"`
	ButtonPath string   `json:"buttonPath,omitempty"`
	Phone      string   `json:"phone,omitempty"`
	Wechat     string   `json:"wechat,omitempty"`
}

type ContentPage struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Slug              string         `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Title             string         `gorm:"size:150;not null" json:"title"`
	Summary           string         `gorm:"size:500" json:"summary"`
	Content           string         `gorm:"type:text" json:"content"`
	BlocksRaw         string         `gorm:"column:blocks;type:longtext" json:"blocksRaw,omitempty"`
	SEOKeywords       string         `gorm:"size:255" json:"seoKeywords"`
	SEOTitle          string         `gorm:"size:180" json:"seoTitle"`
	SEODescription    string         `gorm:"size:500" json:"seoDescription"`
	PageType          string         `gorm:"size:20;not null;default:page;index" json:"pageType"`
	CoverImage        string         `gorm:"size:255" json:"coverImage"`
	AuthorName        string         `gorm:"size:100" json:"authorName"`
	PublishedAt       *time.Time     `gorm:"index" json:"publishedAt,omitempty"`
	RelatedCategoryID uint           `gorm:"index" json:"relatedCategoryId"`
	RelatedProductID  *uint          `gorm:"index" json:"relatedProductId,omitempty"`
	RelatedVendorID   *uint          `gorm:"index" json:"relatedVendorId,omitempty"`
	IsEnabled         bool           `gorm:"default:true;index" json:"isEnabled"`
	PublicationStatus string         `gorm:"size:20;not null;default:published;index" json:"publicationStatus"`
	ContentVersion    uint           `gorm:"not null;default:1" json:"contentVersion"`
	SortOrder         int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p ContentPage) Blocks() []ContentBlock {
	if p.BlocksRaw == "" {
		return nil
	}
	var blocks []ContentBlock
	if err := json.Unmarshal([]byte(p.BlocksRaw), &blocks); err != nil {
		return nil
	}
	return blocks
}

func (p ContentPage) MarshalJSON() ([]byte, error) {
	type Alias ContentPage
	return json.Marshal(struct {
		Alias
		Blocks []ContentBlock `json:"blocks,omitempty"`
	}{Alias: Alias(p), Blocks: p.Blocks()})
}

type FriendLink struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	URL       string         `gorm:"size:255;not null" json:"url"`
	Logo      string         `gorm:"size:255" json:"logo"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	IsEnabled bool           `gorm:"default:true;index" json:"isEnabled"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:100;index" json:"username"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Resource  string    `gorm:"size:100;not null;index" json:"resource"`
	RecordID  uint      `gorm:"index" json:"recordId"`
	Reason    string    `gorm:"size:1000" json:"reason,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
