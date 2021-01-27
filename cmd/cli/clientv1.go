package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
	"io"
	"net/http"
)

type Client struct {
	Scheme string
	Host   string
	Port   int
}

func (c *Client) request(method string, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s://%s:%d%s", c.Scheme, c.Host, c.Port, path)

	var reader io.Reader
	if body != nil {
		out, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(out)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var err *response.BadRequestError
		enc := json.NewDecoder(resp.Body)
		if e := enc.Decode(&err); e != nil {
			return nil, e
		}
		return nil, err
	case http.StatusNotFound:
		return nil, &response.NotFoundError{
			Method: req.Method,
			Path:   req.URL.Path,
		}
	case http.StatusMethodNotAllowed:
		return nil, &response.MethodNotAllowedError{
			Method: req.Method,
			Path:   req.URL.Path,
		}
	case http.StatusInternalServerError:
		return nil, &response.InternalServerError{
			Err: errors.New("unknown server error"),
		}
	default:
		return resp, nil
	}
}

func (c *Client) PutDeployment(id string, req *request.PutDeploymentRequest) (*response.PutDeploymentResponse, *http.Response, error) {
	resp, err := c.request(http.MethodPut, fmt.Sprintf("/api/v1/deployment/%s", id), req)
	if err != nil {
		return nil, resp, err
	}

	var r *response.PutDeploymentResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&r); err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

func (c *Client) DescribeDeployment(id string, req *request.DescribeDeploymentRequest) (*response.DescribeDeploymentResponse, *http.Response, error) {
	resp, err := c.request(http.MethodGet, fmt.Sprintf("/api/v1/deployment/%s", id), nil)
	if err != nil {
		return nil, resp, err
	}

	var r *response.DescribeDeploymentResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&r); err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

func (c *Client) ListDeployments(req *request.ListDeploymentRequest) (*response.ListDeploymentResponse, *http.Response, error) {
	resp, err := c.request(http.MethodGet, "/api/v1/deployment", nil)
	if err != nil {
		return nil, resp, err
	}

	var r *response.ListDeploymentResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&r); err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}
