package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
)

type (
	UserSex      string
	UserStatus   string
	UserAuthType string
	BirthDate    time.Time

	// User оснвная структура ползователя системы
	User struct {
		GeneralTechFields
		AuthType    UserAuthType
		UpdatedAt   time.Time  `json:"-"`
		Email       string     `json:"-"`
		Status      UserStatus `json:"-"`
		VkAuthToken string     `json:"-"`
		VkID        string     `json:"-"`
		Phone       string     `json:"-"`

		// User info
		FirstName *string    `json:"name"`
		Birthday  *BirthDate `json:"birthday"`
		Sex       *UserSex   `json:"sex"`
		City      *string    `json:"city"`
		About     *string    `json:"about"`

		// geo position
		Geolocation geopoint.GeoPoint `json:"-"`

		// Photos
		Photos []UserPhoto `gorm:"foreignKey:UserID" json:"photos"`

		// Matches
		Matches []Match `gorm:"many2many:user_match;" json:"-"`

		// Products
		Products []Product `gorm:"many2many:user_product;" json:"-"`
	}
)

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusNew      UserStatus = "new"

	UserSexMale      UserSex = "male"
	UserSexFemale    UserSex = "female"
	UserSexUndefined UserSex = "undefined"

	UserAuthTypeEmail = "email"
	UserAuthTypeVK    = "vk"
)

// UnmarshalJSON Implement Unmarshalled interface
func (j *BirthDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*j = BirthDate(t)

	return nil
}

// MarshalJSON Implement Marshaller interface
func (j *BirthDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*j).Format("2006-01-02"))
}

func (j *BirthDate) Scan(value interface{}) error {
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("invalid format")
	}
	*j = BirthDate(t)

	return nil
}

func (j *BirthDate) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}

	return time.Time(*j), nil
}

// TableName имя таблицы для gorm
func (u *User) TableName() string {
	return "users"
}
