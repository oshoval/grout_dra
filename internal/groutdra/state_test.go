// SPDX-License-Identifier:Apache-2.0

package groutdra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestRemoveMetadataIfOwned(t *testing.T) {
	const own = `{"apiVersion":"metadata.resource.k8s.io/v1alpha1","metadata":{"uid":"claim-a"}}`
	const ownNext = `{"apiVersion":"metadata.resource.k8s.io/v1beta1","metadata":{"uid":"claim-a"}}`
	const other = `{"apiVersion":"metadata.resource.k8s.io/v1beta1","metadata":{"uid":"claim-b"}}`
	tests := []struct {
		name      string
		data      string
		removed   bool
		wantError bool
	}{
		{name: "one version", data: own, removed: true},
		{name: "two versions", data: own + "\n" + ownNext, removed: true},
		{name: "replacement claim", data: other, removed: false},
		{name: "mixed ownership", data: own + "\n" + other, removed: false},
		{name: "malformed stream", data: own + "\n{", wantError: true},
		{name: "missing owner", data: `{ "metadata": {} }`, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metadata.json")
			require.NoError(t, os.WriteFile(path, []byte(tt.data), 0o600))
			err := removeMetadataIfOwned(types.UID("claim-a"), path)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			_, err = os.Stat(path)
			if tt.removed {
				require.ErrorIs(t, err, os.ErrNotExist)
				return
			}
			require.NoError(t, err)
		})
	}
}
