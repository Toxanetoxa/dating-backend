package entity

import "time"

// Admin is user for admin panel
type Admin struct {
	GeneralTechFields
	Login        string
	PasswordHash string
}

func (a *Admin) TableName() string {
	return "admin"
}

// AdminToken is token for authorize admin
type AdminToken struct {
	GeneralTechFields
	AdminID string
	Secret  string
	Expire  time.Time
}

func (a *AdminToken) TableName() string {
	return "admin_token"
}
