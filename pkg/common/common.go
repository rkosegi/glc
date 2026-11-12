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
	"bytes"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

const (
	DefaultPerPageLimit = 20
)

var (
	ErrTokenRequired    = errors.New("token is required")
	ErrProjectIdMissing = errors.New("project id missing (--project-id)")
	ErrGitDirUnknown    = errors.New("unable to determine project ID from current directory")
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
		return nil, ErrTokenRequired
	}
	// endpoint is not required, client will fall back to https://gitlab.com/api/v4
	if endpoint := d.Endpoint(); endpoint != "" {
		opts = append(opts, gitlab.WithBaseURL(endpoint))
	}
	return gitlab.NewClient(d.Token(), opts...)
}

type ListItemsOpt[T any] func(o *T)

type ItemsFetchFunc[T any] func(d *gitlab.Client, page int64, perPage int64) ([]*T, *gitlab.Response, error)

// FetchItems consumes paged results from gitlab API, possibly capping it to some number.
// Provided fetchFn must correctly update fields Page and PerPage in gitlab.ListOptions with value from the for loop to advance correctly
func FetchItems[T any](gc *gitlab.Client, fetchFn ItemsFetchFunc[T], limit int) ([]*T, error) {
	var (
		err       error
		page      int64
		perPage   int64
		current   []*T
		result    []*T
		remaining int64
		resp      *gitlab.Response
	)

	page = 1
	remaining = -1
	perPage = int64(min(DefaultPerPageLimit, limit))
	for {
		if current, resp, err = fetchFn(gc, page, perPage); err != nil {
			return nil, err
		}
		if remaining == -1 {
			remaining = min(resp.TotalItems, int64(limit)) - int64(len(current))
		}
		page += 1
		if len(current) == 0 {
			break
		}
		result = append(result, current...)
		if limit > 0 && len(result) >= limit {
			break
		}
		// recalc remaining

		if remaining == 0 {
			break
		}
		perPage = min(remaining, perPage)
	}
	return result, err
}

// TODO: maybe just traverse filesystem up to the root and try to find .git/config?
// it might be easier to test and will not be dependent on git binary

func FindGitRemoteUrl(dir string) (string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	c := exec.Command("git", "remote", "-v")
	c.Dir = dir
	c.Stdout = &stdout
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(stdout.String(), "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) == 2 {
			out := parts[1]
			for _, trm := range []string{"fetch", "push"} {
				out = strings.TrimSuffix(out, " ("+trm+")")
			}
			return strings.TrimSpace(out), nil
		}
	}
	return "", nil
}

// ExtractProjectIdFromUrl perform naive extraction of string-based project identifier from URL path.
// This is useful when try to figure out what project ID to use when inside git directory
func ExtractProjectIdFromUrl(projUrl string) string {
	p := strings.TrimPrefix(projUrl, "http://")
	p = strings.TrimPrefix(p, "https://")
	start := strings.Index(p, "/")
	if start == -1 {
		return ""
	}
	p = p[start:]
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, ".git")
	return url.PathEscape(p)
}

// GuessProjectIdFromCwd tries the best effort to get project ID from git URL, if configured.
// No URL is considered as error.
func GuessProjectIdFromCwd() (string, error) {
	u, err := FindGitRemoteUrl(".")
	if err != nil {
		return "", err
	}
	o := ExtractProjectIdFromUrl(u)
	if o == "" {
		return "", ErrGitDirUnknown
	}
	return o, nil
}

func AddProjectIdFlag(pf *pflag.FlagSet, projectId *int64) {
	AddProjectIdFlagWithUsage(pf, projectId, "GitLab project id to use")
}

func AddProjectIdFlagWithUsage(pf *pflag.FlagSet, projectId *int64, usage string) {
	pf.Int64Var(projectId, "project-id", *projectId, usage)
}

func CompressText(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[0:maxLen-4] + "…" + text[len(text)-5:]
	}
	return text
}
