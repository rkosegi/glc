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

package mr

import (
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/rkosegi/glc/pkg/common"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

func StateIs(mrState string) common.ListItemsOpt[gitlab.ListProjectMergeRequestsOptions] {
	return func(lo *gitlab.ListProjectMergeRequestsOptions) {
		lo.State = &mrState
	}
}

func ListInProject(c *gitlab.Client, projId any, maxTotal int, opts ...common.ListItemsOpt[gitlab.ListProjectMergeRequestsOptions]) ([]*gitlab.BasicMergeRequest, error) {
	return common.FetchItems(c, func(cl *gitlab.Client, page int64, perPage int64) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
		lopts := &gitlab.ListProjectMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}
		for _, opt := range opts {
			opt(lopts)
		}
		return cl.MergeRequests.ListProjectMergeRequests(projId, lopts)
	}, maxTotal)
}

type data struct {
	projId         int64
	resolvedProjId any
	gc             *gitlab.Client
	auth           *common.AuthData
	mrs            []*gitlab.BasicMergeRequest
	limit          int
	state          string
}

func formatTime(t *time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func (d *data) ResolveProjID() (err error) {
	if d.projId == 0 {
		d.resolvedProjId, err = common.GuessProjectIdFromCwd()
		return err
	}
	d.resolvedProjId = d.projId
	return nil
}

func newListSubCmd() *cobra.Command {
	d := &data{
		auth:  &common.AuthData{},
		limit: common.DefaultPerPageLimit,
	}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List merge requests",
		Long: `List merge requests in a specific project.

You can specify project using command line flag --project-id or let tool to figure that out
from .git/config file. In the current implementation it uses "git" executable to extract this information
based on current directory.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			if d.gc, err = common.GetGitlabRef(d.auth); err != nil {
				return err
			}

			if err = d.ResolveProjID(); err != nil {
				return err
			}

			opts := make([]common.ListItemsOpt[gitlab.ListProjectMergeRequestsOptions], 0)
			if d.state != "" {
				opts = append(opts, StateIs(d.state))
			}
			if d.mrs, err = ListInProject(d.gc, d.resolvedProjId, d.limit, opts...); err != nil {
				return err
			}

			tbl := tablewriter.NewTable(os.Stdout)
			tbl.Header("", "IID", "Title", "Source", "Target", "Updated", "Author")

			for _, mr := range d.mrs {
				_ = tbl.Append(
					common.LookupMRState2Emoji(mr.State),
					mr.IID,
					common.CompressText(mr.Title, 30),
					common.CompressText(mr.SourceBranch, 25),
					common.CompressText(mr.TargetBranch, 14),
					formatTime(mr.UpdatedAt),
					mr.Author.Username)
			}

			return tbl.Render()
		},
	}

	cmd.Flags().StringVar(&d.state, "state", "", "merge request state. When omitted, all MRs are included")
	cmd.Flags().IntVar(&d.limit, "limit", d.limit, "Maximum number of MRs to list")
	d.auth.AddPFlags(cmd.Flags())
	common.AddProjectIdFlag(cmd.Flags(), &d.projId)
	return cmd
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mr",
		Aliases: []string{"merge-request"},
		Short:   "Merge requests related operations",
	}
	cmd.AddCommand(newListSubCmd())
	return cmd
}
