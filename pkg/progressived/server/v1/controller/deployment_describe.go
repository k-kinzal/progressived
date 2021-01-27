package controller

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// GET /api/v1/deployment/:name
func (c *Controller) DescribeDeployment(name string, _ *request.DescribeDeploymentRequest) (*response.DescribeDeploymentResponse, error) {
	entity, err := c.per.Get(name)
	if err != nil {
		switch err.(type) {
		case *persistence.DeploymentNotFountdError:
			return nil, &response.BadRequestError{Messsage: err.Error()}
		default:
			return nil, &response.InternalServerError{Err: err}
		}
	}

	return response.NewDescribeDeploymentResponseWith(entity), nil
}
