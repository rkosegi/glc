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

package revdiff

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/rkosegi/glc/pkg/common"
	"github.com/samber/lo"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type DiffData struct {
	FromRev       string                      `yaml:"fromRev" json:"fromRev"`
	ToRev         string                      `yaml:"toRev" json:"toRev"`
	UniqueAuthors []string                    `yaml:"uniqueAuthors" json:"uniqueAuthors"`
	MergeRequests []*gitlab.BasicMergeRequest `yaml:"mergeRequests" json:"mergeRequests"`
	Commits       []*gitlab.Commit            `yaml:"commits" json:"commits"`
	Compare       *gitlab.Compare             `yaml:"compare" json:"compare"`
}

type (
	Opt        func(Interface)
	CompareOpt func(opts *CompareSettings)
)

type Interface interface {
	// Compare compares 2 revisions and creates DiffData with all details included, based on provided opts
	Compare(ctx context.Context, from, to string, opts ...CompareOpt) (*DiffData, error)
}

type CompareSettings struct {
	includeMergeCommits bool
}

type differ struct {
	projectRef *gitlab.Project
	gc         *gitlab.Client
	l          *slog.Logger
}

func OptIncludeMergeCommits(include bool) CompareOpt {
	return func(d *CompareSettings) {
		d.includeMergeCommits = include
	}
}

func WithLogger(l *slog.Logger) Opt {
	return func(d Interface) {
		d.(*differ).l = l
	}
}

func (d *differ) init(ctx context.Context, auth *common.AuthData, projectId int) error {
	var err error
	if d.gc, err = common.GetGitlabRef(auth); err != nil {
		return err
	}
	d.l.DebugContext(ctx, "reading project details", "project_id", projectId)
	if d.projectRef, _, err = d.gc.Projects.GetProject(projectId, &gitlab.GetProjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (d *differ) Compare(ctx context.Context, from, to string, opts ...CompareOpt) (*DiffData, error) {
	var (
		err   error
		comp  *gitlab.Compare
		copts CompareSettings
	)

	start := time.Now()
	for _, opt := range opts {
		opt(&copts)
	}
	d.l.DebugContext(ctx, "comparing revisions using API", "project_id", d.projectRef.ID, "from_rev", from, "to_rev", to)
	if comp, _, err = d.gc.Repositories.Compare(d.projectRef.ID, &gitlab.CompareOptions{
		From: &from,
		To:   &to,
	}); err != nil {
		return nil, err
	}

	d.l.DebugContext(ctx, "extracting MRs from commits")
	var (
		mrs     []*gitlab.BasicMergeRequest
		commits []*gitlab.Commit
	)

	for _, commit := range comp.Commits {
		var cmrs []*gitlab.BasicMergeRequest
		d.l.DebugContext(ctx, "extracting MRs associated with commit", "commit_id", commit.ID)
		if cmrs, _, err = d.gc.Commits.ListMergeRequestsByCommit(d.projectRef.ID, commit.ID); err != nil {
			return nil, err
		}
		mrs = append(mrs, cmrs...)

		isMergeCommit := len(commit.ParentIDs) > 1
		if !isMergeCommit || copts.includeMergeCommits {
			commit.Message = strings.Split(commit.Message, "\n")[0]
			commits = append(commits, commit)
		}
	}
	d.l.DebugContext(ctx, "normalizing list of MRs", "count_before", len(mrs))
	mrs = lo.UniqBy(mrs, func(item *gitlab.BasicMergeRequest) int64 {
		return item.IID
	})
	mrs = lo.Filter(mrs, func(item *gitlab.BasicMergeRequest, _ int) bool {
		return item.State == "merged"
	})

	mrsAuthors := lo.Map(mrs, func(item *gitlab.BasicMergeRequest, _ int) string {
		return item.Author.Name
	})
	uniqueCommitAuthors := lo.Map(commits, func(item *gitlab.Commit, _ int) string {
		return item.AuthorName
	})
	uniqueAuthors := lo.Uniq(append(mrsAuthors, uniqueCommitAuthors...))
	comp.Commits = nil

	d.l.DebugContext(ctx, "diff completed", "total_mrs", len(mrs), "total_commits", len(commits),
		"total_authors", len(uniqueAuthors), "time_spent", time.Now().Sub(start).Seconds())
	return &DiffData{
		FromRev:       from,
		ToRev:         to,
		UniqueAuthors: uniqueAuthors,
		MergeRequests: mrs,
		Commits:       commits,
		Compare:       comp,
	}, nil
}

func New(ctx context.Context, auth *common.AuthData, projectId int, opts ...Opt) (Interface, error) {
	d := &differ{}
	for _, opt := range append([]Opt{
		WithLogger(slog.Default()),
	}, opts...) {
		opt(d)
	}
	return d, d.init(ctx, auth, projectId)
}
