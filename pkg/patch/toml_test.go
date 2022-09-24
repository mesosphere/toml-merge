// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOML(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name         string
		ToPatch      string
		Patches      []string
		ExpectError  bool
		ExpectOutput string
	}
	cases := []testCase{
		{
			Name:         "invalid TOML",
			ToPatch:      `🗿`,
			ExpectError:  true,
			ExpectOutput: "",
		},
		{
			Name: "no patches",
			ToPatch: `disabled_plugins = ["restart"]
[plugins.linux]
  shim_debug = true
[plugins.cri.containerd.runtimes.runsc]
  runtime_type = "io.containerd.runsc.v1"`,
			ExpectError: false,
			ExpectOutput: `disabled_plugins = ["restart"]

[plugins]
  [plugins.cri]
    [plugins.cri.containerd]
      [plugins.cri.containerd.runtimes]
        [plugins.cri.containerd.runtimes.runsc]
          runtime_type = "io.containerd.runsc.v1"
  [plugins.linux]
    shim_debug = true
`,
		},
		{
			Name: "invalid patch TOML",
			ToPatch: `disabled_plugins = ["restart"]
[plugins.linux]
  shim_debug = true
[plugins.cri.containerd.runtimes.runsc]
  runtime_type = "io.containerd.runsc.v1"`,
			Patches:     []string{"🏰"},
			ExpectError: true,
		},
		{
			Name: "trivial patch",
			ToPatch: `disabled_plugins = ["restart"]
[plugins.linux]
  shim_debug = true
[plugins.cri.containerd.runtimes.runsc]
  runtime_type = "io.containerd.runsc.v1"`,
			Patches:     []string{`disabled_plugins=[]`},
			ExpectError: false,
			ExpectOutput: `disabled_plugins = []

[plugins]
  [plugins.cri]
    [plugins.cri.containerd]
      [plugins.cri.containerd.runtimes]
        [plugins.cri.containerd.runtimes.runsc]
          runtime_type = "io.containerd.runsc.v1"
  [plugins.linux]
    shim_debug = true
`,
		},
		{
			Name: "patch registry",
			ToPatch: `[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    restrict_oom_score_adj = false
    sandbox_image = "registry.k8s.io/pause:3.7"
    tolerate_missing_hugepages_controller = true
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      discard_unpacked_layers = true
      snapshotter = "overlayfs"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            SystemdCgroup = true
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://anotherregistry:5000"]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler.options]
            SystemdCgroup = true

[proxy_plugins]
  [proxy_plugins.fuse-overlayfs]
    address = "/run/containerd-fuse-overlayfs.sock"
    type = "snapshot"`,
			Patches: []string{`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["https://registry:5000"]`},
			ExpectError: false,
			ExpectOutput: `[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    restrict_oom_score_adj = false
    sandbox_image = "registry.k8s.io/pause:3.7"
    tolerate_missing_hugepages_controller = true
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      discard_unpacked_layers = true
      snapshotter = "overlayfs"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            SystemdCgroup = true
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler.options]
            SystemdCgroup = true
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://registry:5000"]

[proxy_plugins]
  [proxy_plugins.fuse-overlayfs]
    address = "/run/containerd-fuse-overlayfs.sock"
    type = "snapshot"
`,
		},
	}
	for _, tc := range cases {
		tc := tc // capture test case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			out, err := TOMLString(tc.ToPatch, StringPatches(tc.Patches...))
			if tc.ExpectError {
				require.Error(t, err)
			}
			assert.Equal(t, tc.ExpectOutput, out)
		})
	}
}
