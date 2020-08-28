package metrics

import (
	"bytes"
	"text/template"
)

type QueryBuilder struct {
	query string
	data  map[string]interface{}
}

func (qb *QueryBuilder) Build(override map[string]interface{}) (string, error) {
	data := qb.data
	for key, val := range override {
		data[key] = val
	}

	tmpl, err := template.New("query").Parse(qb.query)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func NewQueryBuikder(query string, data map[string]interface{}) *QueryBuilder {
	return &QueryBuilder{
		query: query,
		data:  data,
	}
}
