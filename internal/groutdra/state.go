// SPDX-License-Identifier:Apache-2.0

package groutdra

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/types"
)

// claimState is the durable inventory needed when Unprepare runs in a new
// driver process. It is written before creating any external resources.
type claimState struct {
	UID          types.UID         `json:"uid"`
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	PodClaimName string            `json:"podClaimName,omitempty"`
	Allocations  []allocationState `json:"allocations"`
}

type allocationState struct {
	Request     string `json:"request"`
	Key         string `json:"key"`
	Device      string `json:"device"`
	Pool        string `json:"pool"`
	GatewayCIDR string `json:"gatewayCIDR"`
}

func saveClaimState(paths pathLayout, state claimState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encoding claim state: %w", err)
	}
	if err := writeAtomic(paths.ClaimStatePath(state.UID), data, 0o600); err != nil {
		return fmt.Errorf("saving claim state: %w", err)
	}
	return nil
}

func loadClaimState(paths pathLayout, uid types.UID) (claimState, error) {
	data, err := os.ReadFile(paths.ClaimStatePath(uid))
	if err != nil {
		return claimState{}, err
	}
	var state claimState
	if err := json.Unmarshal(data, &state); err != nil {
		return claimState{}, fmt.Errorf("decoding claim state for %s: %w", uid, err)
	}
	if state.UID != uid {
		return claimState{}, fmt.Errorf("claim state UID %s does not match %s", state.UID, uid)
	}
	return state, nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(mode); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// removeMetadataIfOwned leaves a path alone when a replacement claim has
// already projected its metadata there.
func removeMetadataIfOwned(uid types.UID, path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	// Kubelet can persist multiple API versions as consecutive JSON objects.
	// Check every one before removing a path shared by claims from a template.
	decoder := json.NewDecoder(bytes.NewReader(data))
	seen := false
	for {
		var metadata struct {
			Metadata struct {
				UID types.UID `json:"uid"`
			} `json:"metadata"`
		}
		err := decoder.Decode(&metadata)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading metadata owner %s: %w", path, err)
		}
		if metadata.Metadata.UID == "" {
			return fmt.Errorf("reading metadata owner %s: missing UID", path)
		}
		seen = true
		if metadata.Metadata.UID != uid {
			return nil
		}
	}
	if !seen {
		return fmt.Errorf("reading metadata owner %s: empty file", path)
	}
	return os.Remove(path)
}
