package cmd

import (
	"github.com/k-kinzal/progressived/pkg/progressived"
	"github.com/spf13/cobra"
)

var (
	rollbackCmd = &cobra.Command{
		Use:           "rollback",
		Hidden:        true,
		Short:         "[experimental] Roll back the routing policy",
		RunE:          rollbackRun,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func rollbackRun(*cobra.Command, []string) error {
	pv, err := newProvider(config)
	if err != nil {
		return err
	}

	ms, err := newMetrics(config)
	if err != nil {
		return err
	}

	qb, err := newQueryBuilder(config)
	if err != nil {
		return err
	}

	ag, err := newAlgorithm(config)
	if err != nil {
		return err
	}

	fm, err := newFomura(config)
	if err != nil {
		return err
	}

	p := &progressived.Progressived{
		Provider:    pv,
		Metrics:     ms,
		Builder:     qb,
		Algorithm:   ag,
		Formura:     fm,
		AllowNoData: config.Metrics.AllowNoData,
	}

	if _, err := p.Rollback(); err != nil {
		return err
	}

	return nil
}

func init() {
	rollbackCmd = setFlags(rollbackCmd)
	rootCmd.AddCommand(rollbackCmd)
}
