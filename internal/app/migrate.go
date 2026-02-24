//go:build migrate

package app

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/toxanetoxa/dating-backend/config"

	// migrate tools
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts    = 5
	_defaultTimeout     = time.Second
	SSLModeURLParameter = "sslmode"
)

func init() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	sc := cfg.PG
	var v = make(url.Values)
	v.Set(SSLModeURLParameter, "disable")
	var u = url.URL{
		Scheme:   sc.Db,
		Host:     sc.Host + ":" + strconv.Itoa(sc.Port),
		User:     url.UserPassword(sc.User, sc.Password),
		RawQuery: v.Encode(),
	}
	databaseURL := u.String()

	var (
		attempts = _defaultAttempts
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", databaseURL)
		if err == nil {
			break
		}

		log.Printf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}
