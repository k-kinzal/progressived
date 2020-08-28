package formura

import (
	"github.com/Knetic/govaluate"
)

type Formula struct {
	expression string
}

func (f *Formula) Eval(x float64) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(f.Expression())
	if err != nil {
		return false, err
	}

	parameters := make(map[string]interface{}, 8)
	parameters["x"] = x

	result, err := expression.Evaluate(parameters)
	if err != nil {
		return false, err
	}
	b, ok := result.(bool)
	if !ok {
		return false, err
	}

	return b, nil
}

func (f *Formula) Expression() string {
	return f.expression
}

func NewFormula(expression string) *Formula {
	return &Formula{
		expression: expression,
	}
}
