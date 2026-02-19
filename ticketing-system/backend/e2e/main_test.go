//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := initSharedDB(ctx); err != nil {
		panic("init shared DB: " + err.Error())
	}
	code := m.Run()
	cleanupSharedDB()
	os.Exit(code)
}
