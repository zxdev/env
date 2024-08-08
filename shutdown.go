package env

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Shutdown blocks on context.Context or signal.Notify; only use this
// when env.Graceful is not used; onInterrupt is func that will execute
// when an interrupt if received, when nil just exits
func Shutdown(ctx context.Context, onInterrupt func()) {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-ctx.Done():
	case <-sig:
		signal.Stop(sig)
	}

	if onInterrupt =! nil {
		onInterrupt()
	}
	
	os.Exit(0)

}
