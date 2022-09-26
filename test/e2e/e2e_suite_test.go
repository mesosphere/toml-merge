// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TOML Merge Suite")
}

var testBinary string

var _ = BeforeSuite(func() {
	relPath := filepath.Join(
		"..",
		"..",
		"dist",
		fmt.Sprintf("toml-merge_%s_%s_v1", runtime.GOOS, runtime.GOARCH),
		"toml-merge",
	)
	b, err := exec.LookPath(relPath)
	Expect(err).NotTo(HaveOccurred())
	testBinary = b
})
