package controller

import (
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// GET /api/v1/deployment
func (c *Controller) ListDeployment(_ *request.ListDeploymentRequest) (*response.ListDeploymentResponse, error) {
	entities, err := c.deployments.Seq()
	if err != nil {
		return nil, err
	}

	return response.NewListDeploymentResponse(entities), nil
}
