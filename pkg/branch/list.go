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
	"github.com/rkosegi/glc/pkg/common"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type ListBranchesOpt func(lo *gitlab.ListBranchesOptions)

func OnlyMatchingRe2(re2 string) ListBranchesOpt {
	return func(lo *gitlab.ListBranchesOptions) {
		lo.Regex = &re2
	}
}

func ListWithOpts(c *gitlab.Client, projId any, maxTotal int, opts ...ListBranchesOpt) ([]*gitlab.Branch, error) {
	return common.FetchItems(c, func(cl *gitlab.Client, page int64) ([]*gitlab.Branch, error) {
		lopts := &gitlab.ListBranchesOptions{
			ListOptions: gitlab.ListOptions{
				Page: page,
			},
		}
		for _, opt := range opts {
			opt(lopts)
		}
		branches, _, err := cl.Branches.ListBranches(projId, lopts)
		return branches, err
	}, maxTotal)
}
