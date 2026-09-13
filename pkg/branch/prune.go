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

package branch

import (
	"log/slog"
	"regexp"
	"slices"
	"time"

	"github.com/rkosegi/glc/pkg/common"
	"github.com/rkosegi/glc/pkg/mr"
	xlog "github.com/rkosegi/slog-config"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

func newPruneCmd() *cobra.Command {
	type pruneData struct {
		dryRun       bool
		projectId    int64
		onlyMatching string
		age          int
		includeMR    bool
		gc           *gitlab.Client
		l            *slog.Logger
		projectRef   *gitlab.Project
	}
	xc := xlog.MustNew("info", xlog.LogFormatLogFmt)
	auth := &common.AuthData{}

	d := &pruneData{
		age: 180,
	}
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune branches based on criteria",
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			d.l = xc.Logger()
			ctx := cmd.Context()
			if d.gc, err = common.GetGitlabRef(auth); err != nil {
				return err
			}
			if d.projectId == 0 {
				return common.ErrProjectIdMissing
			}
			d.l.DebugContext(ctx, "reading project details", "project_id", d.projectId)
			if d.projectRef, _, err = d.gc.Projects.GetProject(d.projectId, &gitlab.GetProjectOptions{}); err != nil {
				return err
			}
			if _, err = regexp.Compile(d.onlyMatching); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var mrs []*gitlab.BasicMergeRequest
			if !d.includeMR {
				d.l.DebugContext(cmd.Context(), "listing open project merge requests", "project_id", d.projectId)
				if mrs, err = mr.ListInProject(d.gc, d.projectRef.ID, 500, mr.StateIs("opened")); err != nil {
					return err
				}
				d.l.DebugContext(cmd.Context(), "got MR list", "project_id", d.projectId, "count", len(mrs))
			}
			var opts []ListBranchesOpt
			if d.onlyMatching != "" {
				opts = append(opts, OnlyMatchingRe2(d.onlyMatching))
			}
			d.l.DebugContext(cmd.Context(), "listing branches", "project_id", d.projectId)
			branches, err := ListWithOpts(d.gc, d.projectRef.ID, 100, opts...)
			if err != nil {
				return err
			}
			d.l.DebugContext(cmd.Context(), "got initial branch list", "project_id", d.projectId, "count", len(branches))

			branches = lo.Filter(branches, func(item *gitlab.Branch, _ int) bool {
				isPartOfMR := false
				isRecent := !item.Commit.CommittedDate.Before(time.Now().AddDate(0, 0, -d.age))
				// ignore recent and protected branches
				if isRecent || item.Protected {
					return false
				}

				if !d.includeMR {
					return true
				}

				_, isPartOfMR = lo.Find(mrs, func(mr *gitlab.BasicMergeRequest) bool {
					return mr.SourceBranch == item.Name || mr.TargetBranch == item.Name
				})
				return !isPartOfMR
			})

			slices.SortFunc(branches, func(a, b *gitlab.Branch) int {
				return b.Commit.CommittedDate.Compare(*a.Commit.CommittedDate)
			})

			d.l.InfoContext(cmd.Context(), "final branch list", "project_id", d.projectId, "count", len(branches))

			for _, branch := range branches {
				d.l.WarnContext(cmd.Context(), "removing branch", "project_id", d.projectId, "name", branch.Name,
					"commit_date", branch.Commit.CommittedDate)

				if !d.dryRun {
					if _, err = d.gc.Branches.DeleteBranch(d.projectRef.ID, branch.Name); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	xc.AddPFlags(cmd.Flags())
	auth.AddPFlags(cmd.Flags())
	cmd.Flags().Int64Var(&d.projectId, "project-id", d.projectId, "GitLab project id to get prune branches from. Required")
	cmd.Flags().IntVar(&d.age, "age", d.age, "Age in days, after which is branch considered as stale and will be pruned")
	cmd.Flags().BoolVar(&d.includeMR, "include-mr", false, "Also include branches that has open merge requests")
	cmd.Flags().BoolVar(&d.dryRun, "dry-run", false, "Don't prune any branches, only print what would happen")
	return cmd
}
