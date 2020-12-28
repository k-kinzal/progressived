package cmd

import (
	"context"
	"fmt"
	"github.com/k-kinzal/progressived/cmd/reconciler"
	"github.com/k-kinzal/progressived/pkg/logger"
	"github.com/k-kinzal/progressived/pkg/progressived/conf"
	"github.com/k-kinzal/progressived/pkg/progressived/reconcile"
	"github.com/k-kinzal/progressived/pkg/utils/looper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	rootCmd = &cobra.Command{
		Use:   "progressived",
		Short: "Daemon for progressive delivery",
		PersistentPreRunE: preRun,
		RunE: run,
		Version:       getVersion(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	progressivedConfig *conf.Config
	progressivedLogger logger.Logger
)

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.PersistentFlags().StringP("config", "-c", "/etc/progressived/config.yaml", "config file path")
	rootCmd.PersistentFlags().StringP("format", "o", "json", "log format")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug log")

	rootCmd.AddCommand(reconciler.Command)
}

func preRun(cmd *cobra.Command, _ []string) error {
	configPath, _ := cmd.PersistentFlags().GetString("config")
	config, err := conf.Read(configPath)
	if err != nil {
		return err
	}
	progressivedConfig = config

	format, _ := cmd.PersistentFlags().GetString("format")
	debug, _ := cmd.PersistentFlags().GetBool("debug")
	log := &logrus.Logger{}
	switch format {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	default:
		log.SetFormatter(&logrus.JSONFormatter{})
	}
	if debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	progressivedLogger = log

	return nil
}

func run(_ *cobra.Command, _ []string) error {
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

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("progressived: %v", err))
		os.Exit(1)
	}
	return nil
}



//func newProvider(config *Config) (provider.Provider, error) {
//	var prov provider.Provider
//	switch config.Rule.Provider.Type {
//	case provider.Route53ProviderType:
//		config := &provider.Route53Confg{
//			Sess:                  awsSession,
//			HostedZoneId:          config.Rule.Provider.Route53Provider.HostedZoneId,
//			RecordName:            config.Rule.Provider.Route53Provider.RecordName,
//			SourceIdentifier:      config.Rule.Provider.Route53Provider.SourceIdentifier,
//			DestinationIdentifier: config.Rule.Provider.Route53Provider.DestinationIdentifier,
//		}
//		p, err := provider.NewRoute53Provider(config)
//		if err != nil {
//			return nil, err
//		}
//		prov = p
//	default:
//		panic("this is a bug. please write an issue at https://github.com/k-kinzal/progressived/issues.")
//	}
//
//	return prov, nil
//}
//
//func newMetrics(config *Config) (metrics.Metrics, error) {
//	var met metrics.Metrics
//	switch config.Rule.Metrics.Type {
//	case metrics.CloudWatchMetricsType:
//		config := &metrics.CloudWatchConfig{
//			Sess:   awsSession,
//			Period: config.Rule.Metrics.Period,
//		}
//		met = metrics.NewCloudWatchMetrics(config)
//	default:
//		panic("this is a bug. please write an issue at https://github.com/k-kinzal/progressived/issues.")
//
//	}
//
//	return met, nil
//}
//
//func newAlgorithm(config *Config) (algorithm.Algorithm, error) {
//	var algo algorithm.Algorithm
//	switch config.Rule.Algorithm.Type {
//	case algorithm.IncreaseAlgorithm:
//		algo = algorithm.NewIncretion(config.Rule.Algorithm.Value)
//	case algorithm.DecreaseAlgorithm:
//		algo = algorithm.NewDecrease(config.Rule.Algorithm.Value)
//	default:
//		panic("this is a bug. please write an issue at https://github.com/k-kinzal/progressived/issues.")
//	}
//
//	return algo, nil
//}
//
//func newQueryBuilder(config *Config) (*metrics.QueryBuilder, error) {
//	data := structs.Map(config)
//
//	env := make(map[string]string)
//	for _, v := range os.Environ() {
//		s := strings.Split(v, "=")
//		if len(s) != 2 {
//			continue
//		}
//		env[s[0]] = s[1]
//	}
//	data["Environment"] = env
//
//	return metrics.NewQueryBuikder(config.Rule.Metrics.Query, data), nil
//}
//
//func newFomura(config *Config) (*formura.Formula, error) {
//	return formura.NewFormula(config.Rule.Metrics.Condition), nil
//}
//
//func newProgressived(config *Config) (*progressived.Progressived, error) {
//	pv, err := newProvider(config)
//	if err != nil {
//		return nil, err
//	}
//	mx, err := newMetrics(config)
//	if err != nil {
//		return nil, err
//	}
//	ag, err := newAlgorithm(config)
//	if err != nil {
//		return nil, err
//	}
//	qb, err := newQueryBuilder(config)
//	if err != nil {
//		return nil, err
//	}
//	fm, err := newFomura(config)
//	if err != nil {
//		return nil, err
//	}
//	return &progressived.Progressived{
//		Provider:    pv,
//		Metrics:     mx,
//		Builder:     qb,
//		Algorithm:   ag,
//		Formura:     fm,
//		AllowNoData: config.Rule.Metrics.AllowNoData,
//	}, nil
//}