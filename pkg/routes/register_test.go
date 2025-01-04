package routes

import (
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"git.aetherial.dev/aeth/keiji/pkg/webpages"
	"github.com/gin-gonic/gin"
)

func TestRegister(t *testing.T) {
	e := gin.Default()
	Register(e, "localhost", &storage.SQLiteRepo{}, webpages.FilesystemWebpages{}, auth.EnvAuth{})
}
