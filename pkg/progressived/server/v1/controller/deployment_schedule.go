package controller

import (
	v1 "github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// POST /api/v1/deployment/:name/schedule
func (c *Controller) ScheduleDeployment(name string, req *request.ScheduleDeploymentRequest) (*response.ScheduleDeploymentResponse, error) {
	entity, err := c.deployments.Get(name)
	if err != nil {
		switch err.(type) {
		case *v1.NotFoundError:
			return nil, &response.BadRequestError{Messsage: err.Error()}
		default:
			return nil, &response.InternalServerError{Err: err}
		}
	}

	scheduledEntity, err := entity.Scheduling(req.NextScheduleTime)
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}

	if err := c.deployments.Put(scheduledEntity); err != nil {
		return nil, &response.InternalServerError{Err: err}
	}

	return response.NewScheduleDeploymentResponse(entity), nil
}
