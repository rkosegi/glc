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
	"errors"
	"log/slog"
	"os"

	"github.com/rkosegi/glc/pkg/common"
	xlog "github.com/rkosegi/slog-config"
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var l *slog.Logger
	includeMergeCommits := false
	from := ""
	to := "HEAD"
	projectId := 0
	output := "revdiff.yaml"
	xc := xlog.MustNew("info", xlog.LogFormatLogFmt)
	auth := &common.AuthData{}
	cmd := &cobra.Command{
		Use:   "revdiff",
		Short: "Compute difference between 2 project revisions (tags, commits, ...)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if projectId == 0 {
				return errors.New("project-id is required")
			}
			if to == "" {
				to = "HEAD"
			}
			if from == "" {
				return errors.New("missing from revision (--from)")
			}
			if auth.Token() == "" {
				return errors.New("missing auth token")
			}
			l = xc.Logger()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			rd, err := New(cmd.Context(), auth, projectId, WithLogger(l))
			if err != nil {
				return err
			}
			l.InfoContext(cmd.Context(), "opening output for writing", "output", output)
			var of *os.File
			of, err = os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(0o664))
			if err != nil {
				return err
			}
			defer func(of *os.File) {
				_ = of.Close()
			}(of)
			var out *DiffData
			out, err = rd.Compare(cmd.Context(), from, to, OptIncludeMergeCommits(includeMergeCommits))
			if err != nil {
				return err
			}
			return fluent.TranscodeJson2Yaml[DiffData](out, of)
		},
		Long: `Compute difference between 2 GitLab revisions

In this example, there is previous version tag "v0.0.1" and newly created tag "v0.10.2", which is about to be released.
Following command will generate diff data and writes it to the file "diff.yaml" :

glc revdiff --from v0.0.1 --to v0.10.2 --output diff.yaml
`,
	}

	xc.AddPFlags(cmd.Flags())
	auth.AddPFlags(cmd.Flags())
	cmd.Flags().IntVar(&projectId, "project-id", projectId, "GitLab project id to get revision diff from. Required")
	cmd.Flags().BoolVar(&includeMergeCommits, "include-merge-commits", includeMergeCommits, "Include merge commits in resulting file")
	cmd.Flags().StringVar(&from, "from", from, "Git revision from which to generate diff. Required.")
	cmd.Flags().StringVar(&to, "to", to, "Git revision to which generate diff")
	cmd.Flags().StringVar(&output, "output", output, "File to write the output into. Required.")
	return cmd
}
