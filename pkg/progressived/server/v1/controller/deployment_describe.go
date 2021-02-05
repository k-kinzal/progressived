package controller

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// GET /api/v1/deployment/:name
func (c *Controller) DescribeDeployment(name string, _ *request.DescribeDeploymentRequest) (*response.DescribeDeploymentResponse, error) {
	entity, err := c.deployments.Get(name)
	if err != nil {
		switch err.(type) {
		case *deployment.NotFoundError:
			return nil, &response.BadRequestError{Messsage: err.Error()}
		default:
			return nil, &response.InternalServerError{Err: err}
		}
	}

	return response.NewDescribeDeploymentResponse(entity), nil
}
