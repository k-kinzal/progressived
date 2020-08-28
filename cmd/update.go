package cmd

import (
	"github.com/k-kinzal/progressived/pkg/progressived"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:           "update",
		Hidden:        true,
		Short:         "[experimental] If the metrics match the criteria, the routing policy is updated according to the algorithm",
		RunE:          updateRun,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func updateRun(*cobra.Command, []string) error {
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

	if _, err := p.Update(); err != nil {
		return err
	}

	return nil
}

func init() {
	updateCmd = setFlags(updateCmd)
	rootCmd.AddCommand(updateCmd)
}
