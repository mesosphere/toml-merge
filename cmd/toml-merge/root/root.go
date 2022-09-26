// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mesosphere/dkp-cli-runtime/core/cmd/root"
	"github.com/mesosphere/dkp-cli-runtime/core/output"
)

func NewCommand(in io.Reader, out, errOut io.Writer) (*cobra.Command, output.Output) {
	rootCmd, rootOpts := root.NewCommand(out, errOut)

	o := newOpts(rootOpts.Output)
	o.AddFlags(rootCmd)
	rootCmd.RunE = o.execute

	return rootCmd, rootOpts.Output
}

func Execute() {
	rootCmd, out := NewCommand(os.Stdin, os.Stdout, os.Stderr)
	// disable cobra built-in error printing, we output the error with formatting.
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		out.Error(err, "")
		os.Exit(1)
	}
}
