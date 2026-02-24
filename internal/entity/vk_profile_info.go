package entity

type VKProfileInfo struct {
	ID        string    // id пользователя
	FirstName string    // Имя пользователя.
	Sex       UserSex   // Пол. Возможные значения: 1 — женский,2 — мужской,0 — пол не указан.
	Birthday  BirthDate // Дата рождения пользователя, возвращается в формате D.M.YYYY.
	Email     string    // email пользователя если есть
	Phone     string    // телефон пользователя (должен быть всегда)
}
