package progressived

import "fmt"

func (p *Progressived) Rollback() (float64, error) {
	percentage, err := p.CurrentPercentage()
	if err != nil {
		return -1, fmt.Errorf("rollback: %w", err)
	}
	updatePercentage, err := p.NextPercentage()
	if err != nil {
		return -1, fmt.Errorf("rollback: %w", err)
	}
	if percentage == updatePercentage {
		return -1, AlreadyCompletedError{}
	}

	if err := p.Provider.Update(updatePercentage); err != nil {
		return -1, fmt.Errorf("rollback: %w", err)
	}

	return updatePercentage, nil
}
