package circleci

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

const (
	queryLimit = 100 // maximum that CircleCI allows
)

var (
	defaultBaseURL = &url.URL{Host: "circleci.com", Scheme: "https", Path: "/api/v1.1/"}
	defaultLogger  = log.New(os.Stderr, "", log.LstdFlags)
)

// Logger is a minimal interface for injecting custom logging logic for debug logs
type Logger interface {
	Printf(fmt string, args ...interface{})
}

// APIError represents an error from CircleCI
type APIError struct {
	HTTPStatusCode int
	Message        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.HTTPStatusCode, e.Message)
}

type ApiClient struct {
	BaseURL    *url.URL     // CircleCI API endpoint (defaults to DefaultEndpoint)
	Token      string       // CircleCI API token (needed for private repositories and mutative actions)
	HTTPClient *http.Client // HTTPClient to use for connecting to CircleCI (defaults to http.DefaultClient)

	Debug  bool   // debug logging enabled
	Logger Logger // logger to send debug messages on (if enabled), defaults to logging to stderr with the standard flags
}

func (c *ApiClient) baseURL() *url.URL {
	if c.BaseURL == nil {
		return defaultBaseURL
	}

	return c.BaseURL
}

func (c *ApiClient) client() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}

	return c.HTTPClient
}

func (c *ApiClient) logger() Logger {
	if c.Logger == nil {
		return defaultLogger
	}

	return c.Logger
}

func (c *ApiClient) debug(format string, args ...interface{}) {
	if c.Debug {
		c.logger().Printf(format, args...)
	}
}

func (c *ApiClient) debugRequest(req *http.Request) {
	if c.Debug {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			c.debug("error debugging request %+v: %s", req, err)
		}
		c.debug("request:\n%+v", string(out))
	}
}

func (c *ApiClient) debugResponse(resp *http.Response) {
	if c.Debug {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			c.debug("error debugging response %+v: %s", resp, err)
		}
		c.debug("response:\n%+v", string(out))
	}
}

// FollowProject follows a project
func (c *ApiClient) FollowProject(vcstype, account, reponame string) (*Project, error) {
	response := &Project{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/follow", vcstype, account, reponame), response, nil, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ListProjects returns the list of projects the user is watching
func (c *ApiClient) ListProjects() ([]*Project, error) {
	projects := []*Project{}

	err := c.request("GET", "projects", &projects, nil, nil)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject retrieves a specific project
// Returns nil of the project is not in the list of watched projects
func (c *ApiClient) GetProject(vcstype, account, reponame string) (*Project, error) {
	projects, err := c.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if vcstype == project.VcsType && account == project.Username && reponame == project.Reponame {
			return project, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Unable to find project %s/%s/%s", vcstype, account, reponame))
}

// DisableProject disables a project
func (c *ApiClient) DisableProject(vcstype, account, reponame string) error {
	return c.request("DELETE", fmt.Sprintf("project/%s/%s/%s/enable", vcstype, account, reponame), nil, nil, nil)
}

// SetAwsKeys updates projects AWS keys
func (c *ApiClient) SetAwsKeys(vcstype, account, reponame, keyId, secret string) error {
	return c.request("PUT", fmt.Sprintf("project/%s/%s/%s/settings", vcstype, account, reponame), nil, nil, &Project{
		AWSConfig: AWSConfig{
			AWSKeypair: &AWSKeypair{
				AccessKey: keyId,
				SecretKey: secret,
			},
		},
	})
}

// RemoveAwsKeys updates projects AWS keys
func (c *ApiClient) RemoveAwsKeys(vcstype, account, reponame string) error {
	return c.request("PUT", fmt.Sprintf("project/%s/%s/%s/settings", vcstype, account, reponame), nil, nil, &Project{
		AWSConfig: AWSConfig{
			AWSKeypair: nil,
		},
	})
}

// ListEnvVars list environment variable to the specified project
// Returns the env vars (the value will be masked)
func (c *ApiClient) ListEnvVars(vcstype, account, reponame string) ([]EnvVar, error) {
	envVar := []EnvVar{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/envvar", vcstype, account, reponame), &envVar, nil, nil)
	if err != nil {
		return nil, err
	}

	return envVar, nil
}

// AddEnvVar adds a new environment variable to the specified project
// Returns the added env var (the value will be masked)
func (c *ApiClient) AddEnvVar(vcstype, account, reponame, name, value string) (*EnvVar, error) {
	envVar := &EnvVar{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/envvar", vcstype, account, reponame), envVar, nil, &EnvVar{Name: name, Value: value})
	if err != nil {
		return nil, err
	}

	return envVar, nil
}

// DeleteEnvVar deletes the specified environment variable from the project
func (c *ApiClient) DeleteEnvVar(vcstype, account, reponame, name string) error {
	return c.request("DELETE", fmt.Sprintf("project/%s/%s/%s/envvar/%s", vcstype, account, reponame, name), nil, nil, nil)
}

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

func (c *ApiClient) request(method, path string, responseStruct interface{}, params url.Values, bodyStruct interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("circle-token", c.Token)

	u := c.baseURL().ResolveReference(&url.URL{Path: path, RawQuery: params.Encode()})

	c.debug("building request for %s", u)

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return err
	}

	if bodyStruct != nil {
		b, err := json.Marshal(bodyStruct)
		if err != nil {
			return err
		}

		req.Body = nopCloser{bytes.NewBuffer(b)}
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	c.debugRequest(req)

	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.debugResponse(resp)

	if resp.StatusCode >= 300 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &APIError{HTTPStatusCode: resp.StatusCode, Message: "unable to parse response: %s"}
		}

		if len(body) > 0 {
			message := struct {
				Message string `json:"message"`
			}{}
			err = json.Unmarshal(body, &message)
			if err != nil {
				return &APIError{
					HTTPStatusCode: resp.StatusCode,
					Message:        fmt.Sprintf("unable to parse API response: %s", err),
				}
			}
			return &APIError{HTTPStatusCode: resp.StatusCode, Message: message.Message}
		}

		return &APIError{HTTPStatusCode: resp.StatusCode}
	}

	if responseStruct != nil {
		err = json.NewDecoder(resp.Body).Decode(responseStruct)
		if err != nil {
			return err
		}
	}

	return nil
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AWSConfig represents AWS configuration for a project
type AWSConfig struct {
	AWSKeypair *AWSKeypair `json:"keypair"`
}

// AWSKeypair represents the AWS access/secret key for a project
// SecretKey will be a masked value
type AWSKeypair struct {
	AccessKey string `json:"access_key_id"`
	SecretKey string `json:"secret_access_key"`
}

type Project struct {
	AWSConfig AWSConfig `json:"aws"`
	Username  string    `json:"username"`
	Reponame  string    `json:"reponame"`
	VcsType   string    `json:"vcs_type"`
}
