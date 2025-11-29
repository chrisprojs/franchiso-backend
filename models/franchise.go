package models

import (
	"time"

	"github.com/google/uuid"
)

type Franchise struct {
	tableName       struct{}  `pg:"franchiso.franchises"`
	ID              uuid.UUID `pg:"id" json:"id"`
	UserID          uuid.UUID `pg:"user_id" json:"user_id"`
	CategoryID      uuid.UUID `pg:"category_id" json:"category_id"`
	Brand           string    `pg:"brand" json:"brand"`
	Logo            string    `pg:"logo" json:"logo"`
	AdPhotos        []string  `pg:"ad_photos,array" json:"ad_photos"`
	Description     string    `pg:"description" json:"description"`
	Investment      int       `pg:"investment" json:"investment"`
	MonthlyRevenue  int       `pg:"monthly_revenue" json:"monthly_revenue"`
	ROI             int       `pg:"roi" json:"roi"`
	BranchCount     int       `pg:"branch_count" json:"branch_count"`
	YearFounded     int       `pg:"year_founded" json:"year_founded"`
	Website         string    `pg:"website" json:"website"`
	WhatsappContact string    `pg:"whatsapp_contact" json:"whatsapp_contact"`
	IsBoosted       bool      `pg:"is_boosted,use_zero" json:"is_boosted"`
	Stpw            string    `pg:"stpw" json:"stpw"`
	NIB             string    `pg:"nib" json:"nib"`
	NPWP            string    `pg:"npwp" json:"npwp"`
	Status          string    `pg:"status" json:"status"`
	CreatedAt       time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt       time.Time `pg:"updated_at" json:"updated_at"`

	User     *User     `pg:"rel:has-one,fk:user_id" json:"user"`
	Category *Category `pg:"rel:has-one,fk:category_id" json:"category"`
}

type FranchiseES struct {
	ID              string     `json:"id"`
	User            UserES     `json:"user"`
	Category        CategoryES `json:"category"`
	Brand           string     `json:"brand"`
	Logo            string     `json:"logo"`
	AdPhotos        []string   `json:"ad_photos"`
	Description     string     `json:"description"`
	Investment      int        `json:"investment"`
	MonthlyRevenue  int        `json:"monthly_revenue"`
	ROI             int        `json:"roi"`
	BranchCount     int        `json:"branch_count"`
	YearFounded     int        `json:"year_founded"`
	Website         string     `json:"website"`
	WhatsappContact string     `json:"whatsapp_contact"`
	IsBoosted       bool       `json:"is_boosted"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

type UserES struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
}

type CategoryES struct {
	CategoryID string `json:"category_id"`
	Category   string `json:"category"`
}
