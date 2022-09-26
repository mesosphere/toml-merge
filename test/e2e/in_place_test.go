// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
)

var _ = Describe("in place editing", func() {
	var tempDir string

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
	})

	DescribeTable(
		"success",
		func(inputFile, expectedOutputFile string, patches ...string) {
			fmt.Println(inputFile, expectedOutputFile, patches)
			expectedOutputFilePath := filepath.Join("testdata", "outputs", expectedOutputFile)
			writtenOutputFilePath := filepath.Join(tempDir, filepath.Base(expectedOutputFile))
			if os.Getenv("CREATE_TEST_DATA") == "true" {
				writtenOutputFilePath = expectedOutputFilePath
			}
			Expect(
				copy.Copy(filepath.Join("testdata", "inputs", inputFile), writtenOutputFilePath),
			).To(Succeed())
			for i := range patches {
				patches[i] = filepath.Join("testdata", "inputs", patches[i])
			}

			cmd := exec.Command( //nolint:gosec // Running the binary in e2e tests is fine.
				testBinary,
				"-i",
				"--patch-file",
				strings.Join(patches, ","),
				writtenOutputFilePath,
			)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "cmd output: %s", output)

			expectedOutput, err := os.ReadFile(expectedOutputFilePath)
			Expect(err).NotTo(HaveOccurred())
			actualOutput, err := os.ReadFile(writtenOutputFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actualOutput)).To(Equal(string(expectedOutput)))
		},

		Entry(
			"single file patch",
			"containerd_default_config.toml",
			"single_file_patch.toml",
			"mirror_auth_patch.toml",
		),

		Entry(
			"multiple file patches",
			"containerd_default_config.toml", "multiple_file_patches.toml",
			"mirror_auth_patch.toml", "mirror_endpoint_patch.toml",
		),

		Entry(
			"patch globs",
			"containerd_default_config.toml", "multiple_file_patches.toml",
			"*_patch.toml",
		),
	)
})
