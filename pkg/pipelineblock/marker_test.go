package pipelineblock_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smykla-labs/klaudiush/pkg/pipelineblock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_SetBlockMarker(t *testing.T) {
	t.Run("creates marker file when enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(markerPath)
		require.NoError(t, err)
	})

	t.Run("does nothing when disabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, false)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		// Verify file does not exist
		_, err = os.Stat(markerPath)
		require.True(t, os.IsNotExist(err))
	})
}

func TestManager_CheckBlockMarker(t *testing.T) {
	t.Run("returns valid marker when not expired", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		marker, valid := mgr.CheckBlockMarker()
		assert.True(t, valid)
		assert.NotNil(t, marker)
		assert.Equal(t, "GIT019", marker.ErrorCode)
		assert.Equal(t, "validate-git-push", marker.Validator)
		assert.Equal(t, "Force push blocked", marker.Message)
	})

	t.Run("returns false when marker expired", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 1*time.Millisecond, true)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		marker, valid := mgr.CheckBlockMarker()
		assert.False(t, valid)
		assert.Nil(t, marker)

		// Verify marker file was cleaned up
		_, err = os.Stat(markerPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("returns false when no marker exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		marker, valid := mgr.CheckBlockMarker()
		assert.False(t, valid)
		assert.Nil(t, marker)
	})

	t.Run("clears marker when env var set", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		// Set clear env var
		t.Setenv("KLAUDIUSH_CLEAR_PIPELINE_BLOCK", "1")

		marker, valid := mgr.CheckBlockMarker()
		assert.False(t, valid)
		assert.Nil(t, marker)

		// Verify marker file was cleaned up
		_, err = os.Stat(markerPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		// Create marker with enabled manager
		enabledMgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)
		err := enabledMgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		// Check with disabled manager
		disabledMgr := pipelineblock.NewManager(markerPath, 5*time.Second, false)
		marker, valid := disabledMgr.CheckBlockMarker()
		assert.False(t, valid)
		assert.Nil(t, marker)
	})
}

func TestManager_ClearBlockMarker(t *testing.T) {
	t.Run("removes marker file", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		err := mgr.SetBlockMarker("GIT019", "validate-git-push", "Force push blocked")
		require.NoError(t, err)

		err = mgr.ClearBlockMarker()
		require.NoError(t, err)

		// Verify file removed
		_, err = os.Stat(markerPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("does not error when marker does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, true)

		err := mgr.ClearBlockMarker()
		require.NoError(t, err)
	})

	t.Run("does nothing when disabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		markerPath := filepath.Join(tmpDir, "block-marker")

		// Create marker manually
		err := os.WriteFile(markerPath, []byte("test"), 0600)
		require.NoError(t, err)

		mgr := pipelineblock.NewManager(markerPath, 5*time.Second, false)

		err = mgr.ClearBlockMarker()
		require.NoError(t, err)

		// Verify file still exists (disabled manager doesn't clear)
		_, err = os.Stat(markerPath)
		require.NoError(t, err)
	})
}

func TestManager_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		mgr := pipelineblock.NewManager("/tmp/marker", 5*time.Second, true)
		assert.True(t, mgr.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		mgr := pipelineblock.NewManager("/tmp/marker", 5*time.Second, false)
		assert.False(t, mgr.IsEnabled())
	})
}
