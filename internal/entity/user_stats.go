package entity

type SexStats struct {
	Male      int64 `json:"male"`
	Female    int64 `json:"female"`
	Undefined int64 `json:"undefined"`
}
type UsersStats struct {
	TotalUsers    int64    `json:"totalUsers"`
	UsersByDay    int64    `json:"usersByDay"`
	UsersByWeek   int64    `json:"usersByWeek"`
	UsersActive   int64    `json:"usersActive"`
	UsersNew      int64    `json:"usersNew"`
	UsersInActive int64    `json:"usersInActive"`
	SexStats      SexStats `json:"sexStats"`
}
