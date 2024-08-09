package env

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Shutdown blocks on context.Context or signal.Notify; only use this
// when env.Graceful is not used; interrupt() is func that will execute
// when an interrupt is received before exiting, when nil just exits
func Shutdown(ctx context.Context, interrupt func()) {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-ctx.Done():
	case <-sig:
		signal.Stop(sig)
	}

	if interrupt != nil {
		interrupt()
	}

	os.Exit(0)

}
