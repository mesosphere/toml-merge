// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"fmt"
	"os"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/jimmidyson/toml-merge/pkg/patch"

	"github.com/mesosphere/dkp-cli-runtime/core/output"
)

type opts struct {
	patchFileGlobs []string
	inPlace        bool
	backupSuffix   string
	out            output.Output
}

func newOpts(out output.Output) *opts {
	return &opts{out: out}
}

func (o *opts) AddFlags(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(1)
	cmd.Flags().StringSliceVarP(&o.patchFileGlobs, "patch-file", "p", nil,
		"patch files to apply, also accepts glob patterns")
	_ = cmd.MarkFlagRequired("patch-file")
	cmd.Flags().BoolVarP(&o.inPlace, "in-place", "i", false, "edit files in place")
	cmd.Flags().StringVarP(&o.backupSuffix, "backup-suffix", "b", "",
		"create a backup with specified suffix if edit --in-place is specified")
	cmd.Flags().Lookup("backup-suffix").NoOptDefVal = ".bak"
}

func (o *opts) execute(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	patched, err := patch.TOMLFile(inputFile, patch.FileGlobPatches(o.patchFileGlobs...))
	if err != nil {
		return err
	}
	if !o.inPlace {
		_, err = o.out.ResultWriter().Write([]byte(patched))
		return err
	}

	if o.backupSuffix != "" {
		if err := copy.Copy(inputFile, fmt.Sprintf("%s%s", inputFile, o.backupSuffix),
			copy.Options{
				Sync:          true,
				PreserveTimes: true,
				PreserveOwner: true,
			}); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	if err := os.WriteFile(
		inputFile,
		[]byte(patched),
		0o600,
	); err != nil {
		return fmt.Errorf("failed to overwrite file with patched config: %w", err)
	}
	return nil
}
