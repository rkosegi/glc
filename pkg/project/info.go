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

package project

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/rkosegi/glc/pkg/common"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type data struct {
	projId         string
	auth           *common.AuthData
	gc             *gitlab.Client
	resolvedProjId any
}

func doInfo(d *data) error {
	ap, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	pu, err := common.FindGitRemoteUrl(ap)
	if err != nil {
		return fmt.Errorf("unable to extract git info from directory: %w", err)
	}
	if pu == "" {
		return fmt.Errorf("no remote url found")
	}
	fmt.Printf("URL: %s\n", pu)

	d.projId = common.ExtractProjectIdFromUrl(pu)
	fmt.Printf("ID: %s\n", url.PathEscape(d.projId))

	if d.gc, err = common.GetGitlabRef(d.auth); err != nil {
		return err
	}

	proj, _, err := d.gc.Projects.GetProject(d.projId, &gitlab.GetProjectOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Description: %v\n", proj.Description)
	fmt.Printf("Created: %v\n", proj.CreatedAt)

	return nil
}

func newInfoSubCommand() *cobra.Command {
	d := &data{auth: &common.AuthData{}}
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display project information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return doInfo(d)
		},
	}

	common.AddProjectIdFlag(cmd.Flags(), &d.projId)
	return cmd
}
