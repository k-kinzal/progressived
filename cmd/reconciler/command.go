package reconciler

import (
	"context"
	"github.com/k-kinzal/progressived/pkg/progressived/reconcile"
	"github.com/k-kinzal/progressived/pkg/utils/looper"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)


var (
	Command = &cobra.Command{
		Use:           "reconciler",
		Short:         "",
		RunE:          run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	// aguments
	interval int
)

func init() {
	Command.Flags().IntVar(&interval, "interval", 0, "")
}

func run(*cobra.Command, []string) error {
	ctx := context.Background()

	loop := looper.New(runtime.NumCPU())
	go loop.Run(time.Duration(interval) * time.Second, reconcile.Func)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case s := <- sigs:
			if s == syscall.SIGTERM && interval > 0 {
				ctx, _ := context.WithTimeout(ctx, time.Duration(interval+5) * time.Second)
				return loop.Shutdown(ctx)
			}
			return nil
		}
	}
}


