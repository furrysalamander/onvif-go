package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/furrysalamander/onvif-go/internal/mockcam"
)

func WithMockServer(t *testing.T, fn func(addr string)) {
	t.Helper()

	mc := mockcam.New()
	go func() {
		mc.Listen()
	}()

	time.Sleep(50 * time.Millisecond)

	if mc.Addr == "" {
		t.Fatal("mockcam address is empty")
	}

	fn(mc.Addr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mc.Shutdown(ctx); err != nil {
		t.Logf("mockcam shutdown: %v", err)
	}
}

var _ = fmt.Sprintf("")
