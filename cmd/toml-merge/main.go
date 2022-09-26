// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"math/rand"
	"time"

	"github.com/jimmidyson/toml-merge/cmd/toml-merge/root"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	root.Execute()
}
