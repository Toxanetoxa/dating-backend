package dto

import (
	"fmt"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/vkapi"
)

func VKAPIProfileInfoToVKProfileInfo(vk vkapi.ProfileInfo) (*entity.VKProfileInfo, error) {
	info := entity.VKProfileInfo{
		ID:        vk.ID,
		FirstName: vk.FirstName,
		Email:     vk.Email,
		Phone:     vk.Phone,
	}

	sexes := map[int]entity.UserSex{
		1: entity.UserSexFemale,
		2: entity.UserSexMale,
		0: entity.UserSexUndefined,
	}

	info.Sex = sexes[vk.Sex]

	bDay, err := time.Parse("02.01.2006", vk.Birthday)
	if err != nil {
		return nil, fmt.Errorf("invalid birthday: %w", err)
	}

	info.Birthday = entity.BirthDate(bDay)

	return &info, nil
}
