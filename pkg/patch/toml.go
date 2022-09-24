// Copyright 2022 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	burntoml "github.com/BurntSushi/toml"
	jsonpatch "github.com/evanphx/json-patch/v5"
	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v3"
)

type PatchesFunc func() ([]*toml.Tree, error)

func StringPatches(patches ...string) PatchesFunc {
	return func() ([]*toml.Tree, error) {
		trees := make([]*toml.Tree, 0, len(patches))
		for _, patch := range patches {
			tree, err := toml.LoadBytes([]byte(patch))
			if err != nil {
				return nil, fmt.Errorf("failed to load toml patch: %w", err)
			}
			trees = append(trees, tree)
		}

		return trees, nil
	}
}

func FilePatches(patchFiles ...string) PatchesFunc {
	return func() ([]*toml.Tree, error) {
		trees := make([]*toml.Tree, 0, len(patchFiles))
		for _, patch := range patchFiles {
			tree, err := toml.LoadFile(patch)
			if err != nil {
				return nil, fmt.Errorf("failed to load toml patch: %w", err)
			}
			trees = append(trees, tree)
		}

		return trees, nil
	}
}

func ReaderPatches(patchReaders ...io.Reader) PatchesFunc {
	return func() ([]*toml.Tree, error) {
		trees := make([]*toml.Tree, 0, len(patchReaders))
		for _, patch := range patchReaders {
			tree, err := toml.LoadReader(patch)
			if err != nil {
				return nil, fmt.Errorf("failed to load toml patch: %w", err)
			}
			trees = append(trees, tree)
		}

		return trees, nil
	}
}

func CombinePatches(pFns ...PatchesFunc) PatchesFunc {
	return func() ([]*toml.Tree, error) {
		var trees []*toml.Tree

		for _, p := range pFns {
			t, err := p()
			if err != nil {
				return nil, err
			}
			trees = append(trees, t...)
		}

		return trees, nil
	}
}

// TOMLString patches a TOML string toPatch with the patches (should be TOML merge patches).
func TOMLString(
	toPatch string,
	p PatchesFunc,
) (string, error) { // we use github.com.pelletier/go-toml here to unmarshal arbitrary TOML to JSON
	tree, err := toml.LoadBytes([]byte(toPatch))
	if err != nil {
		return "", fmt.Errorf("failed to load original toml: %w", err)
	}

	return applyTOMLPatches(tree, p)
}

// TOMLFile patches TOML from a file with the patches (should be TOML merge patches).
func TOMLFile(
	toPatchFile string,
	p PatchesFunc,
) (string, error) { // we use github.com.pelletier/go-toml here to unmarshal arbitrary TOML to JSON
	tree, err := toml.LoadFile(toPatchFile)
	if err != nil {
		return "", fmt.Errorf("failed to load original toml: %w", err)
	}

	return applyTOMLPatches(tree, p)
}

// TOMLReader patches TOML from a reader with the patches (should be TOML merge patches).
func TOMLReader(
	toPatchReader io.Reader,
	p PatchesFunc,
) (string, error) { // we use github.com.pelletier/go-toml here to unmarshal arbitrary TOML to JSON
	tree, err := toml.LoadReader(toPatchReader)
	if err != nil {
		return "", fmt.Errorf("failed to load original toml: %w", err)
	}

	return applyTOMLPatches(tree, p)
}

func applyTOMLPatches(tree *toml.Tree, p PatchesFunc) (string, error) {
	patches, err := p()
	if err != nil {
		return "", fmt.Errorf("failed to load toml patches: %w", err)
	}

	// convert to JSON for patching
	j, err := tomlToJSON(tree)
	if err != nil {
		return "", fmt.Errorf("failed to convert original toml to json: %w", err)
	}
	// apply merge patches
	for _, patch := range patches {
		pj, err := tomlToJSON(patch)
		if err != nil {
			return "", fmt.Errorf("failed to convert toml patch to json: %w", err)
		}
		patched, err := jsonpatch.MergePatch(j, pj)
		if err != nil {
			return "", fmt.Errorf("failed to apply toml patch: %w", err)
		}
		j = patched
	}
	// convert result back to TOML
	return jsonToTOMLString(j)
}

// tomlToJSON converts arbitrary TOML to JSON.
func tomlToJSON(tree *toml.Tree) ([]byte, error) {
	b, err := json.Marshal(tree.ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to toml: %w", err)
	}
	return b, nil
}

// jsonToTOMLString converts arbitrary JSON to TOML.
func jsonToTOMLString(j []byte) (string, error) {
	var unstruct interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshalling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	if err := yaml.Unmarshal(j, &unstruct); err != nil {
		return "", fmt.Errorf("failed to unmarshal json: %w", err)
	}
	// we use github.com/BurntSushi/toml here because github.com.pelletier/go-toml
	// can only marshal structs AND BurntSushi/toml is what contained uses
	// and has more canonically formatted output (we initially plan to use
	// this package for patching containerd config)
	var buff bytes.Buffer
	if err := burntoml.NewEncoder(&buff).Encode(unstruct); err != nil {
		return "", fmt.Errorf("failed to encode to toml: %w", err)
	}
	return buff.String(), nil
}
