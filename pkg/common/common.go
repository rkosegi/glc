/*
Copyright 2026 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"errors"
	"os"

	"github.com/spf13/pflag"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type AuthData struct {
	EndpointEnvironmentVariable string
	ResolvedEndpoint            string `yaml:"-" json:"-"`

	TokenEnvironmentVariable string
	ResolvedToken            string `yaml:"-" json:"-"`
}

func (d *AuthData) Token() string {
	if d.TokenEnvironmentVariable == "" {
		d.TokenEnvironmentVariable = "GITLAB_TOKEN"
	}
	if d.ResolvedToken == "" {
		d.ResolvedToken = os.Getenv(d.TokenEnvironmentVariable)
	}
	return d.ResolvedToken
}

func (d *AuthData) Endpoint() string {
	if d.EndpointEnvironmentVariable == "" {
		d.EndpointEnvironmentVariable = "CI_API_V4_URL"
	}
	if d.ResolvedEndpoint == "" {
		d.ResolvedEndpoint = os.Getenv(d.EndpointEnvironmentVariable)
	}
	return d.ResolvedEndpoint
}

func (d *AuthData) AddPFlags(cmd *pflag.FlagSet) {
	cmd.StringVar(&d.TokenEnvironmentVariable, "gitlab-token-env", "GITLAB_TOKEN",
		"Name of the environment variable with GitLab API token. Required")
	cmd.StringVar(&d.EndpointEnvironmentVariable, "gitlab-endpoint-env", "CI_API_V4_URL",
		"Name of the environment variable with GitLab API endpoint. Optional")
}

func GetGitlabRef(d *AuthData, opts ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	token := d.Token()
	if token == "" {
		return nil, errors.New("token is required")
	}
	// endpoint is not required, client will fall back to https://gitlab.com/api/v4
	if endpoint := d.Endpoint(); endpoint != "" {
		opts = append(opts, gitlab.WithBaseURL(endpoint))
	}
	return gitlab.NewClient(d.Token(), opts...)
}
