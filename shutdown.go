package env

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Shutdown waits on control context.Context or signal.Notify; only use this
// when env.Graceful is not used; shutdownFunc will execute after a system or user
// signal is received (can be nil), however when a context.CancelFunc acutally
// needs to be called before exiting (or anything else for control purposes)
// then pass these items wrapped as the shutdownFunc; uses os.Exit(0)
//
//	ctx, cancel:= context.WithCancel(context.Backgroud())
//	env.Shutdown(ctx, func(){cancel()})
func Shutdown(ctx context.Context, shutdownFunc func()) {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-ctx.Done(): // program flow signal
	case <-sig: // system interrupt or user sighup|sigterm signal
		signal.Stop(sig) // got a signal; one is enough
	}

	if shutdownFunc != nil {
		shutdownFunc()
	}

	os.Exit(0)

}
