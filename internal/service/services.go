package service

// Services менеджер сервисов приложения
type Services struct {
	Profile        ProfileService
	Auth           AuthService
	Find           FindService
	Like           LikeService
	Match          MatchService
	Chat           ChatService
	AdminAuth      AdminAuthService
	AdminUsers     AdminUsersService
	PaymentService PaymentService
	AdminProfile   AdminProfileService
}

func NewServices(u ProfileService, a AuthService, f FindService, l LikeService, m MatchService, c ChatService, au AdminAuthService, us AdminUsersService, ps PaymentService, aps AdminProfileService) *Services {
	return &Services{
		Profile:        u,
		Auth:           a,
		Find:           f,
		Like:           l,
		Match:          m,
		Chat:           c,
		AdminAuth:      au,
		AdminUsers:     us,
		PaymentService: ps,
		AdminProfile:   aps,
	}
}
