package cmd

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/controller"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// server option
	serverPort int

	rootCmd = &cobra.Command{
		Use:               "progressived",
		Short:             "Daemon for progressive delivery",
		PersistentPreRunE: prepare,
		RunE:              run,
		Version:           GetVersion(),
		SilenceErrors:     true,
		SilenceUsage:      true,
	}
)

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.PersistentFlags().StringP("format", "o", "text", "log format")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug log")
	rootCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "")
}

func prepare(_ *cobra.Command, _ []string) error {
	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	serverLogger := &logrus.Logger{Out: os.Stderr}
	format, _ := cmd.PersistentFlags().GetString("format")
	switch format {
	case "json":
		serverLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	default:
		serverLogger.SetFormatter(&logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}
	debug, _ := cmd.PersistentFlags().GetBool("debug")
	if debug {
		serverLogger.SetLevel(logrus.DebugLevel)
	} else {
		serverLogger.SetLevel(logrus.InfoLevel)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGUSR2)

	ctrl := controller.NewController(deployment.NewFactory(), deployment.NewDeploymentsOnInMemory(), serverLogger)

	errCh := make(chan error, 1)
	handler := &http.ServeMux{}
	handler.HandleFunc("/", ctrl.Handler)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", serverPort), Handler: handler}
	go func() {
		serverLogger.Debugf("listen to the server on `%s`", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case s := <-sigs:
			var ctx context.Context
			if s == syscall.SIGTERM || s == syscall.SIGUSR2 {
				serverLogger.Debug("start the shutdown gracefully")
				ctx = context.Background()
			} else {
				serverLogger.Debug("start the shutdown")
				ctx, _ = context.WithTimeout(context.Background(), time.Second)
			}
			if err := srv.Shutdown(ctx); err != nil {
				errCh <- err
			}
		case err := <-errCh:
			switch err.Error() {
			case "http: Server closed":
				serverLogger.Debug("closed successfully")
				return nil
			case "context deadline exceeded":
				serverLogger.Warn("the shutdown did not complete successfully. processing of the in-flight request may have been interrupted. if you want it to finish successfully, send `SIGTERM`")
				return nil
			default:
				return err
			}
		}
	}
}

func Execute() error {
	return rootCmd.Execute()
}
