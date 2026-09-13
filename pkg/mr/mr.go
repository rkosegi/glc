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
	"github.com/rkosegi/glc/pkg/common"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type ListInProjectOpt func(lo *gitlab.ListProjectMergeRequestsOptions)

func StateIs(mrState string) ListInProjectOpt {
	return func(lo *gitlab.ListProjectMergeRequestsOptions) {
		lo.State = &mrState
	}
}

func ListInProject(c *gitlab.Client, projId any, maxTotal int, opts ...ListInProjectOpt) ([]*gitlab.BasicMergeRequest, error) {
	return common.FetchItems(c, func(cl *gitlab.Client, page int64) ([]*gitlab.BasicMergeRequest, error) {
		lopts := &gitlab.ListProjectMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				Page: page,
			},
		}
		for _, opt := range opts {
			opt(lopts)
		}
		mrs, _, err := cl.MergeRequests.ListProjectMergeRequests(projId, lopts)
		return mrs, err
	}, maxTotal)
}
