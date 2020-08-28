package progressived

import "fmt"

func (p *Progressived) Rollback() (float64, error) {
	percentage, err := p.Provider.Get()
	if err != nil {
		return -1, fmt.Errorf("rollback: %w", err)
	}
	updatePercentage := p.Algorithm.Previous(percentage)
	if percentage <= 0 && updatePercentage <= 0 {
		return -1, fmt.Errorf("rollback: progress of rollback is already complete")
	}
	if percentage >= 100 && updatePercentage >= 100 {
		return -1, fmt.Errorf("rollback: progress of rollback is already complete")
	}
	if updatePercentage <= 0 {
		updatePercentage = 0
	}
	if updatePercentage >= 100 {
		updatePercentage = 100
	}
	if err := p.Provider.Update(updatePercentage); err != nil {
		return -1, fmt.Errorf("rollback: %w", err)
	}

	return updatePercentage, nil
}
