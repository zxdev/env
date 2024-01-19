package env

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Shutdown blocks on context.Context or signal.Notify; only use this
// when env.Graceful is not used; exit will call os.Exit(0)
func Shutdown(ctx context.Context, exit bool) {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-ctx.Done():
	case <-sig:
		signal.Stop(sig)
	}

	if exit {
		os.Exit(0)
	}
}
