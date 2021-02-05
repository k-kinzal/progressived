package controller

import (
	"encoding/json"
	"fmt"
	"github.com/k-kinzal/progressived/pkg/logger"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
	"net/http"
	"regexp"
)

var (
	prefix                             = "/api/v1"
	regexpPathDeployment               = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/?$`, prefix))
	regexpPathDeploymentWithId         = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/([a-z][a-z0-9_]*/?$)`, prefix))
	regexpPathDeploymentScheduleWithId = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/([a-z][a-z0-9_]*/schedule$)`, prefix))
	regexpPathDeploymentPauseWithId    = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/([a-z][a-z0-9_]*/pause$)`, prefix))
)

type Controller struct {
	deployments deployment.Deployments
	logger      logger.Logger
}

func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}

	if name := regexpPathDeploymentScheduleWithId.FindString(r.URL.Path); name != "" {
		switch r.Method {
		case http.MethodPost:
			var req *request.ScheduleDeploymentRequest
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&req); err != nil {
				c.ErrorHandler(w, r, err)
				return
			}

			res, err := c.ScheduleDeployment(name, req)
			if err != nil {
				c.ErrorHandler(w, r, err)
				return
			}
			resp = res
		default:
			c.ErrorHandler(w, r, &response.MethodNotAllowedError{Method: r.Method, Path: r.URL.Path})
		}
	} else if name := regexpPathDeploymentPauseWithId.FindString(r.URL.Path); name != "" {
		switch r.Method {
		case http.MethodPost:
			res, err := c.PauseDeployment(name, &request.PauseDeploymentRequest{})
			if err != nil {
				c.ErrorHandler(w, r, err)
				return
			}
			resp = res
		default:
			c.ErrorHandler(w, r, &response.MethodNotAllowedError{Method: r.Method, Path: r.URL.Path})
		}
	} else if name := regexpPathDeploymentWithId.FindString(r.URL.Path); name != "" {
		switch r.Method {
		case http.MethodGet:
			res, err := c.DescribeDeployment(name, &request.DescribeDeploymentRequest{})
			if err != nil {
				c.ErrorHandler(w, r, err)
				return
			}
			resp = res
		case http.MethodPut:
			var req *request.PutDeploymentRequest
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&req); err != nil {
				c.ErrorHandler(w, r, err)
				return
			}

			res, err := c.PutDeployment(name, req)
			if err != nil {
				c.ErrorHandler(w, r, err)
				return
			}
			resp = res
		default:
			c.ErrorHandler(w, r, &response.MethodNotAllowedError{Method: r.Method, Path: r.URL.Path})
			return
		}
	} else if regexpPathDeployment.MatchString(r.URL.Path) {
		switch r.Method {
		case http.MethodGet:
			res, err := c.ListDeployment(&request.ListDeploymentRequest{})
			if err != nil {
				c.ErrorHandler(w, r, err)
				return
			}
			resp = res
		default:
			c.ErrorHandler(w, r, &response.MethodNotAllowedError{Method: r.Method, Path: r.URL.Path})
			return
		}
	} else {
		c.ErrorHandler(w, r, &response.NotFoundError{Method: r.Method, Path: r.URL.Path})
		return
	}

	out, err := json.Marshal(resp)
	if err != nil {
		c.ErrorHandler(w, r, &response.InternalServerError{Err: err})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(out)
}

func (c *Controller) ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch e := err.(type) {
	// 4xx
	case *response.BadRequestError:
		out, err := json.Marshal(e)
		if err != nil {
			c.ErrorHandler(w, r, &response.InternalServerError{Err: err})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)

	case *response.NotFoundError:
		w.WriteHeader(http.StatusNotFound)
	case *response.MethodNotAllowedError:
		w.WriteHeader(http.StatusMethodNotAllowed)
	// 5xx
	case *response.InternalServerError:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		panic(fmt.Sprintf("unknown error: open the issue from https://github.com/k-kinzal/progressived/issues: %v", err))
	}
}

func NewController(deployments deployment.Deployments, logger logger.Logger) *Controller {
	return &Controller{
		deployments: deployments,
		logger:      logger,
	}
}
