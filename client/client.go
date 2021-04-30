package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiKey     string
	projectId  string
	httpClient *http.Client
}

type ContentType struct {
	Id       string                   `json:"id"`
	Name     string                   `json:"name"`
	CodeName string                   `json:"codename"`
	Elements []map[string]interface{} `json:"elements"`
}

type ContentTypes struct {
	Types []ContentType `json:"types"`
}

func NewClient(apiKey string, projectId string) *Client {
	return &Client{
		apiKey:     apiKey,
		projectId:  projectId,
		httpClient: &http.Client{},
	}
}

func (c *Client) GetContentTypes() (*ContentTypes, error) {
	body, err := c.httpRequest("types", "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	contentTypes := ContentTypes{}
	err = json.NewDecoder(body).Decode(&contentTypes)
	if err != nil {
		return nil, err
	}
	return &contentTypes, nil
}

func (c *Client) NewContentType(contentType *ContentType) (string, error) {
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(contentType)
	if err != nil {
		return "", err
	}
	response, err := c.httpRequest("types", "POST", buf)
	if err != nil {
		return "", err
	}

	contentTypeResponse := ContentType{}
	err = json.NewDecoder(response).Decode(&contentTypeResponse)

	if err != nil {
		return "", err
	}

	return contentTypeResponse.Id, nil
}

func (c *Client) GetContentType(id string) (*ContentType, error) {
	body, err := c.httpRequest(fmt.Sprintf("types/%v", id), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	contentType := &ContentType{}
	err = json.NewDecoder(body).Decode(contentType)
	if err != nil {
		return nil, err
	}
	return contentType, nil
}

func (c *Client) UpdateContentType(contentType *ContentType) error {
	err := c.DeleteContentType(contentType.Id)
	if err != nil {
		return err
	}

	response, err := c.NewContentType(contentType)
	if err != nil {
		return err
	}

	contentType.Id = response

	return nil
}

func (c *Client) DeleteContentType(id string) error {
	_, err := c.httpRequest(fmt.Sprintf("types/%s", id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) requestPath(path string) string {
	return fmt.Sprintf("https://manage.kontent.ai/v2/projects/%s/%s", c.projectId, path)
}

func (c *Client) httpRequest(path, method string, body bytes.Buffer) (closer io.ReadCloser, err error) {
	req, err := http.NewRequest(method, c.requestPath(path), &body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	switch method {
	case "GET":
	case "DELETE":
	default:
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		respBody := new(bytes.Buffer)
		_, err := respBody.ReadFrom(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("got a non 200 status code: %v", resp.StatusCode)
		}
		return nil, fmt.Errorf("got a non 200 status code: %v - %s", resp.StatusCode, respBody.String())
	}
	return resp.Body, nil
}
