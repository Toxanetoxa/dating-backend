package integration_tests

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	host                = "app:8032"
	healthPath          = "/health"
	healthCheckAttempts = 20

	smockerAdminPath = "http://partner-mock:8081"

	basePath = "http://" + host + "/api/v1"
)

//go:embed mocks.yaml
var mocks string

func TestMain(m *testing.M) {
	err := healthCheck(basePath+healthPath, healthCheckAttempts)
	if err != nil {
		log.Fatalf("host %s is not available", host)
	}

	log.Printf("Integration tests: host %s is available", host)

	err = healthCheck(smockerAdminPath+"/version", healthCheckAttempts)
	if err != nil {
		log.Fatalf("Integration tests: url %s is not available: %s", smockerAdminPath, err)
	}

	log.Printf("Integration tests: url %s is available", smockerAdminPath)

	err = Do(
		Post(smockerAdminPath+"/mocks"),
		Send().Headers("Content-Type").Add("application/x-yaml"),
		Send().Body().String(mocks),
		Expect().Status().Equal(http.StatusOK),
	)
	if err != nil {
		log.Fatalf("could not add mocks: %v", err)
	}

	time.Sleep(time.Second) // ждём на всякий случай, чтобы мок сохранился

	os.Exit(m.Run())
}

func healthCheck(url string, attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(url), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", url, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return err
}
