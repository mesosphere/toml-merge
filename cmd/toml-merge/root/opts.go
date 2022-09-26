// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/jimmidyson/toml-merge/pkg/patch"

	"github.com/mesosphere/dkp-cli-runtime/core/output"
)

type opts struct {
	inputFile      string
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
	patched, err := patch.TOMLFile(o.inputFile, patch.FileGlobPatches(o.patchFileGlobs...))
	if err != nil {
		return err
	}
	if !o.inPlace {
		_, err = o.out.ResultWriter().Write([]byte(patched))
		return err
	}

	f, err := os.CreateTemp(
		filepath.Dir(o.inputFile),
		fmt.Sprintf(".%s.tmp", filepath.Base(o.inputFile)),
	)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := f.WriteString(patched); err != nil {
		return fmt.Errorf("failed to write patched TOML to temporary file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if o.backupSuffix != "" {
		if err := copy.Copy(o.inputFile, fmt.Sprintf("%s%s", o.inputFile, o.backupSuffix),
			copy.Options{
				Sync:          true,
				PreserveTimes: true,
				PreserveOwner: true,
			}); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	if err := os.Rename(f.Name(), o.inputFile); err != nil {
		return fmt.Errorf("failed to overwrite file with patched file")
	}
	return nil
}
