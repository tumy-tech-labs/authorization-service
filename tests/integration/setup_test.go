package integration

import (
	"os"
	"testing"

	"github.com/bradtumy/authorization-service/internal/middleware"
)

func TestMain(m *testing.M) {
	os.Setenv("OIDC_CONFIG_FILE", "/dev/null")
	middleware.LoadOIDCConfig()
	os.Exit(m.Run())
}
