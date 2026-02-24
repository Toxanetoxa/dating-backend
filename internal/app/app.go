package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/toxanetoxa/dating-backend/config"
	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/internal/service/repo"
	v1 "github.com/toxanetoxa/dating-backend/internal/transport/http/v1"
	"github.com/toxanetoxa/dating-backend/internal/vkapi"
	"github.com/toxanetoxa/dating-backend/internal/youkassa"
	"github.com/toxanetoxa/dating-backend/pkg/logging"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Run(config *config.Config) error {
	// logger
	l := logging.InitLogger(config.Log.Level)

	// metrics
	m := entity.NewMetrics()
	prometheus.MustRegister(m.LoginTotal)
	prometheus.MustRegister(m.LogoutTotal)
	prometheus.MustRegister(m.DeleteUserTotal)
	prometheus.MustRegister(m.RegistrationTotal)
	prometheus.MustRegister(m.RequestsTotal)
	prometheus.MustRegister(m.Requests20x)
	prometheus.MustRegister(m.Requests40x)
	prometheus.MustRegister(m.Requests50x)
	prometheus.MustRegister(m.BuyBoostTotal)
	prometheus.MustRegister(m.SuccessPayments)
	prometheus.MustRegister(m.CanceledPayments)
	prometheus.MustRegister(m.FailedPayments)

	// postgres gorm
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.PG.Host,
		config.PG.User,
		config.PG.Password,
		config.PG.Db,
		config.PG.Port,
		config.PG.SSL,
	)

	dbLog := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logger.Error, // todo can turn on for debug (move to app config)
	})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError:         true,
		SkipDefaultTransaction: true,
		Logger:                 dbLog,
	})
	if err != nil {
		return fmt.Errorf("could not connect database: %w", err)
	}

	// rdb tokens
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	// s3
	s3Client, err := minio.New(config.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.S3.AccessKey, config.S3.SecretKey, ""),
		Secure: config.S3.Secure,
	})
	if err != nil {
		return fmt.Errorf("could not connect s3: %w", err)
	}

	// FCM
	fcmService, err := service.NewPushFirebaseService(l, config.Firebase.Enable, config.Firebase.JsonCredentialsPath, repo.NewUserRepo(db, l))
	if err != nil {
		return fmt.Errorf("could not connect to Firebase: %w", err)
	}

	// YKassa client
	ykClient := youkassa.NewYouKassaClient(config.YouKassa.BaseURL, config.YouKassa.ShopID, config.YouKassa.SecretKey)

	//Vk api client
	vkAPIClient := vkapi.NewVKAPIClient(config.VK.BaseURL, config.VK.AppId)

	matchRepo := repo.NewMatchRepoPg(db, l)

	// init services (use cases)
	emailService := service.NewEmailService(l, config.Smtp.Enable, config.Smtp.Host, config.Smtp.Port, config.Smtp.User, config.Smtp.Password, config.Smtp.From)
	profileService := service.NewProfileService(l, repo.NewUserRepo(db, l), repo.NewUserPhotoRepo(db, l), s3Client, config.S3.BucketName, redisClient, config.S3.BaseFilesURL, m)
	authService := service.NewAuthService(redisClient, emailService, profileService, l, config.Auth.TokenTTL, config.Auth.AuthCodeTTL, !config.Smtp.Enable, m, vkAPIClient)
	findService := service.NewFindService(l, repo.NewUserRepo(db, l), matchRepo, fcmService, repo.NewPaymentsRepo(db, l))
	likeService := service.NewLikeService(l, repo.NewUserRepo(db, l))
	matchService := service.NewMatchService(l, matchRepo)
	chatService := service.NewChatService(l, repo.NewMatchRepoPg(db, l), repo.NewChatRepoPg(db, l), fcmService)

	// admin''s services (need to move)
	adminAuthService := service.NewAdminAuthService(l, repo.NewAdminRepoPg(db, l))
	adminUsersService := service.NewAdminUsers(l, repo.NewUserRepo(db, l))
	adminProfileService := service.NewAdminProfileService(l, repo.NewAdminRepoPg(db, l))

	paymentService := service.NewPaymentService(l, repo.NewPaymentsRepo(db, l), repo.NewUserRepo(db, l), ykClient, config.YouKassa.RedirectUrl, fcmService, m)

	// service manager
	services := service.NewServices(profileService, authService, findService, likeService, matchService, chatService, adminAuthService, adminUsersService, paymentService, adminProfileService)

	// create server
	s := v1.NewEchoServer(services, l, m, config.Log.BodyDump)

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// start server
	go func() {
		l.Info("starting server", "port", config.HTTP.Port)
		if err := s.Start(fmt.Sprintf(":%d", config.HTTP.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("shutting down the server", "err", err.Error())
		}
	}()

	// wait for interrupt signal to gracefully shut down the server with a timeout of second
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		l.Error(err.Error())
	}

	return nil
}
