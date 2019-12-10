package observe

import (
	"context"
	//"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/machinebox/graphql"
)

var (
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrUnauthorized    = errors.New("authorization error")
)

type Dataset struct {
	WorkspaceID string     `json:"workspaceId"`
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Transform   *Transform `json:"transform"`
}

type backendDataset struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Transform struct {
		Current struct {
			Stages []struct {
				Pipeline string  `json:"pipeline"`
				Input    []Input `json:"input"`
			} `json:"stages"`
		} `json:"current"`
	} `json:"transform"`
}

// Transform is simplified - we support only one stage
type Transform struct {
	Inputs   []Input  `json:"inputs"`
	Pipeline []string `json:"pipeline"`
}

type CreateDatasetInput struct {
	WorkspaceID string   `json:"workspaceId"`
	Label       string   `json:"label"`
	Inputs      []Input  `json:"inputs"`
	Pipeline    []string `json:"pipeline"`
}

type Input struct {
	InputName string `json:"inputName"`
	DatasetID string `json:"datasetId"`
}

type authTripper struct {
	http.RoundTripper
	key string
}

func SanitizePipeline(p string) (result []string) {
	for _, line := range strings.Split(strings.TrimSpace(p), "\n") {
		for _, stmt := range strings.Split(line, "|") {
			result = append(result, strings.TrimSpace(stmt))
		}
	}
	return result
}

func (t *authTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.key != "" {
		req.Header.Set("Authorization", "Bearer "+t.key)
	}

	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}
	s, _ := httputil.DumpRequest(req, true)
	log.Printf("%s\n", s)

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	s, _ = httputil.DumpResponse(resp, true)
	log.Printf("%s\n", s)

	switch resp.StatusCode {
	case http.StatusOK:
		return resp, err
	case http.StatusUnprocessableEntity:
		s, _ := httputil.DumpResponse(resp, true)
		panic(string(s))
		return resp, err
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, fmt.Errorf("received unexpected status code %d", resp.StatusCode)
	}
}

type Client struct {
	client *graphql.Client
}

// Verify checks if we can connect to API.
func (c *Client) Verify() error {
	req := graphql.NewRequest(`
	{
		currentUser {
			id
		}
	}`)

	var respData struct {
		Response struct {
			Id string `json:"id"`
		} `json:"currentUser"`
	}

	if err := c.client.Run(context.Background(), req, &respData); err != nil {
		return err
	}

	return nil
}

func convertDataset(d *backendDataset) (*Dataset, error) {
	if d == nil {
		return nil, ErrDatasetNotFound
	}

	var t *Transform

	switch len(d.Transform.Current.Stages) {
	case 0:
	case 1:
		stage := d.Transform.Current.Stages[0]
		t = &Transform{
			Pipeline: SanitizePipeline(stage.Pipeline),
			Inputs:   stage.Input,
		}
	default:
		return nil, fmt.Errorf("unsupported transform, more than one stage defined")
	}

	dataset := &Dataset{
		WorkspaceID: "1", // hack
		ID:          d.ID,
		Label:       d.Label,
		Transform:   t,
	}

	return dataset, nil
}

func (c *Client) CreateDataset(input CreateDatasetInput) (*Dataset, error) {
	req := graphql.NewRequest(`
mutation ($workspaceId: ObjectId!, $datasetId: ObjectId, $label: String!, $pipeline: String!, $inputs: [InputDefinitionInput!]!) {
	  saveDataset(
	  workspaceId:$workspaceId
	  dataset: {
		id: $datasetId
		label: $label
		deleted: false
	  }
	  transform: {
		outputStage: "0"
		stages: [{
		  stageID: "0"
		  pipeline: $pipeline
		  input: $inputs
		}]
	  }) {
		dataset {
		  id
		  label
		  transform {
			id
			current {
			  stages {
				pipeline
			  }
			}
		  }
		}
	  }
	}`)

	req.Var("workspaceId", input.WorkspaceID)
	req.Var("label", input.Label)
	req.Var("pipeline", strings.Join(input.Pipeline, " | "))
	req.Var("inputs", input.Inputs)

	var respData struct {
		Response struct {
			Dataset *backendDataset `json:"dataset"`
		} `json:"saveDataset"`
	}

	if err := c.client.Run(context.Background(), req, &respData); err != nil {
		return nil, err
	}

	return convertDataset(respData.Response.Dataset)
}

func (c *Client) LookupDataset(workspaceID string, label string) (*Dataset, error) {
	// TODO: we need an endpoint to lookup dataset by label
	req := graphql.NewRequest(`
	query ($workspaceId: ObjectId!) {
	  project(projectId:$workspaceId) {
		datasets {
			id
		  label
		}
	  }
	}`)

	req.Var("workspaceId", workspaceID)

	var respData struct {
		Project struct {
			Datasets []struct {
				ID    string `json:"id"`
				Label string `json:"label"`
			} `json:"datasets"`
		} `json:"project"`
	}

	if err := c.client.Run(context.Background(), req, &respData); err != nil {
		return nil, err
	}

	for _, d := range respData.Project.Datasets {
		if d.Label == label {
			return c.GetDataset(d.ID)
		}
	}

	return nil, ErrDatasetNotFound
}

func (c *Client) GetDataset(id string) (*Dataset, error) {
	req := graphql.NewRequest(`
	query GetDataset($id: ObjectId!) {
		dataset(id:$id) {
			id
			label
			transform {
				current {
					stages {
						pipeline
						input {
							inputName
							datasetId
						}
					}
				}
			}
		}
	}`)

	req.Var("id", id)

	var respData struct {
		Dataset *backendDataset `json:"dataset"`
	}

	if err := c.client.Run(context.Background(), req, &respData); err != nil {
		return nil, err
	}

	return convertDataset(respData.Dataset)
}

func (c *Client) DeleteDataset(id string) error {
	req := graphql.NewRequest(`
	mutation ($id: ObjectId!) {
		deleteDataset(dsid: $id) {
			success
			errorMessage
		}
	}`)

	req.Var("id", id)
	var respData struct {
		Success bool `json:"success"`
	}

	if err := c.client.Run(context.Background(), req, &respData); err != nil {
		return err
	}

	return nil
}

func NewClient(baseURL string, key string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	authed := &http.Client{
		Transport: &authTripper{key: key},
	}

	c := &Client{
		client: graphql.NewClient(u.String(), graphql.WithHTTPClient(authed)),
	}

	return c, c.Verify()
}
