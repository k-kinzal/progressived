package controller

import (
	v1 "github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// POST /api/v1/deployment/:name/pause
func (c *Controller) PauseDeployment(name string, _ *request.PauseDeploymentRequest) (*response.PauseDeploymentResponse, error) {
	entity, err := c.deployments.Get(name)
	if err != nil {
		switch err.(type) {
		case *v1.NotFoundError:
			return nil, &response.BadRequestError{Messsage: err.Error()}
		default:
			return nil, &response.InternalServerError{Err: err}
		}
	}

	scheduledEntity, err := entity.Pause()
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}

	if err := c.deployments.Put(scheduledEntity); err != nil {
		return nil, &response.InternalServerError{Err: err}
	}

	return response.NewPauseDeploymentResponse(entity), nil
}
