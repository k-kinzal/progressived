package controller

import (
	"encoding/json"
	"fmt"
	"github.com/k-kinzal/progressived/pkg/logger"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
	"net/http"
	"regexp"
)

var (
	prefix                     = "/api/v1"
	regexpPathDeployment       = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/?$`, prefix))
	regexpPathDeploymentWithId = regexp.MustCompile(fmt.Sprintf(`^%s/deployment/([0-9]+/?$)`, prefix))
)

type Controller struct {
	fact   persistence.DeploymentFactory
	per    persistence.Deployments
	logger logger.Logger
}

func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}

	if id := regexpPathDeploymentWithId.FindString(r.URL.Path); id != "" {
		switch r.Method {
		case http.MethodGet:
			res, err := c.DescribeDeployment(id, &request.DescribeDeploymentRequest{})
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

			res, err := c.PutDeployment(id, req)
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

func NewController(fact persistence.DeploymentFactory, per persistence.Deployments, logger logger.Logger) *Controller {
	return &Controller{
		fact:   fact,
		per:    per,
		logger: logger,
	}
}
