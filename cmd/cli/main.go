package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/toxanetoxa/dating-backend/config"
	"github.com/toxanetoxa/dating-backend/internal/entity"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// postgres gorm
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.PG.Host,
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Db,
		cfg.PG.Port,
		cfg.PG.SSL,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError:         true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("could not connect database: %v", err)
	}

	app := &cli.App{
		Name:  "dating-cli",
		Usage: "useful tool for dating admins !",
		Action: func(*cli.Context) error {
			fmt.Println("hello, use help flag to get commands list")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "create-admin",
				Aliases:     []string{"ca"},
				Description: "Create admin user",
				Usage:       "ca login password",
				Action: func(c *cli.Context) error {
					log.Println("Create admin user:", c.Args().Get(0))
					admin := entity.Admin{
						GeneralTechFields: entity.GeneralTechFields{ID: uuid.New().String()},
						Login:             c.Args().Get(0),
						PasswordHash:      hashPassword(c.Args().Get(1)),
					}
					err = db.WithContext(c.Context).Model(&entity.Admin{}).Create(&admin).Error
					if err != nil {
						log.Println("can't create admin:", err)

						return cli.Exit("fail", 1)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// cratch
func hashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass + "!%:R_adf45674567"))

	return fmt.Sprintf("%x", h.Sum(nil))
}
