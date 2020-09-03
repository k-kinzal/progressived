package progressived

func (p *Progressived) TargetName() string {
	return p.Provider.TargetName()
}

func (p *Progressived) CurrentPercentage() (float64, error) {
	pct, err := p.Provider.Get()
	if err != nil {
		return -1, err
	}
	return pct, nil
}

func (p *Progressived) NextPercentage() (float64, error) {
	pct, err := p.CurrentPercentage()
	if err != nil {
		return -1, err
	}
	npct := p.Algorithm.Next(pct)
	if npct <= 0 {
		npct = 0
	}
	if npct >= 100 {
		npct = 100
	}
	return npct, nil
}

func (p *Progressived) PreviousPercentage() (float64, error) {
	pct, err := p.CurrentPercentage()
	if err != nil {
		return -1, err
	}
	ppct := p.Algorithm.Previous(pct)
	if ppct <= 0 {
		ppct = 0
	}
	if ppct >= 100 {
		ppct = 100
	}
	return ppct, nil
}
