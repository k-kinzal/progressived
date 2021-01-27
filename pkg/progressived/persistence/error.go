package persistence

import "fmt"

type DeploymentNotFountdError struct {
	Name string
}

func (e *DeploymentNotFountdError) Error() string {
	return fmt.Sprintf("%s: not found", e.Name)
}
