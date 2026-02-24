package dto

import (
	"github.com/toxanetoxa/dating-backend/internal/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toxanetoxa/dating-backend/internal/vkapi"
)

func TestVKAPIProfileInfoToVKProfileInfo(t *testing.T) {
	t.Run("sunny case", func(t *testing.T) {

		// потом сделать табличные

		testDate, err := time.Parse("02.01.2006", "29.07.1999")

		if err != nil {
			t.Fatalf("could not parse date")
		}

		data := &entity.VKProfileInfo{
			ID:        "159659261",
			FirstName: "Tester",
			Sex:       entity.UserSexMale,
			Birthday:  entity.BirthDate(testDate),
			Email:     "",
			Phone:     "375345610109",
		}

		res, err := VKAPIProfileInfoToVKProfileInfo(vkapi.ProfileInfo{
			ID:        "159659261",
			FirstName: "Tester",
			LastName:  "Testerov",
			Sex:       2,
			Birthday:  "29.07.1999",
			Email:     "",
			Phone:     "375345610109",
			Verified:  false,
		})

		assert.NoError(t, err)
		assert.Equal(t, data, res)

	})
}
