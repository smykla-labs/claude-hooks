package backup

import (
	"os"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/smykla-labs/klaudiush/pkg/config"
)

var (
	// ErrConfigFileNotFound is returned when config file doesn't exist.
	ErrConfigFileNotFound = errors.New("config file not found")

	// ErrBackupDisabled is returned when backup system is disabled.
	ErrBackupDisabled = errors.New("backup system is disabled")
)

// Manager orchestrates backup operations.
type Manager struct {
	// storage provides persistence for snapshots.
	storage Storage

	// config contains backup configuration.
	config *config.BackupConfig
}

// NewManager creates a new backup manager.
func NewManager(storage Storage, cfg *config.BackupConfig) (*Manager, error) {
	if storage == nil {
		return nil, errors.New("storage cannot be nil")
	}

	if cfg == nil {
		cfg = &config.BackupConfig{}
	}

	return &Manager{
		storage: storage,
		config:  cfg,
	}, nil
}

// CreateBackupOptions contains options for creating a backup.
type CreateBackupOptions struct {
	// ConfigPath is the absolute path to the config file.
	ConfigPath string

	// Trigger indicates what caused this backup.
	Trigger Trigger

	// Metadata provides additional context.
	Metadata SnapshotMetadata
}

// CreateBackup creates a new backup snapshot with deduplication.
func (m *Manager) CreateBackup(opts CreateBackupOptions) (*Snapshot, error) {
	if !m.config.IsEnabled() {
		return nil, ErrBackupDisabled
	}

	// Read config file
	data, err := os.ReadFile(opts.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(ErrConfigFileNotFound, opts.ConfigPath)
		}

		return nil, errors.Wrap(err, "failed to read config file")
	}

	// Initialize storage if needed
	if !m.storage.Exists() {
		if initErr := m.storage.Initialize(); initErr != nil {
			return nil, errors.Wrap(initErr, "failed to initialize storage")
		}
	}

	// Load index
	index, err := m.storage.LoadIndex()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load index")
	}

	// Compute content hash
	contentHash := ComputeContentHash(data)
	opts.Metadata.ConfigHash = contentHash

	// Deduplication: Check if identical content already exists
	if existing, found := index.FindByHash(contentHash); found {
		return &existing, nil
	}

	// Determine storage type (full vs patch)
	timestamp := time.Now()
	storageType := m.determineStorageType(index)

	// Generate snapshot ID
	snapshotID := GenerateSnapshotID(timestamp, contentHash)

	// Determine sequence number and chain ID
	chainID := m.generateChainID(index)
	seqNum := m.getNextSequenceNumber(index, chainID)

	// For now, only implement full snapshots (patch support in Phase 3)
	var storagePath string

	var size int64

	if storageType != StorageTypeFull {
		// Patch support will be implemented in Phase 3
		return nil, errors.New("patch snapshots not yet implemented")
	}

	storagePath, err = m.storage.Save(snapshotID+".full.toml", data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save full snapshot")
	}

	size = int64(len(data))

	// Determine config type
	configType := m.determineConfigType(opts.ConfigPath)

	// Create snapshot
	snapshot := Snapshot{
		ID:             snapshotID,
		SequenceNum:    seqNum,
		Timestamp:      timestamp,
		ConfigPath:     opts.ConfigPath,
		ConfigType:     configType,
		Trigger:        opts.Trigger,
		StorageType:    storageType,
		StoragePath:    storagePath,
		Size:           size,
		Checksum:       contentHash,
		ChainID:        chainID,
		BaseSnapshotID: "",
		PatchFrom:      "",
		Metadata:       opts.Metadata,
	}

	// Add to index
	index.Add(snapshot)

	// Save index
	if err := m.storage.SaveIndex(index); err != nil {
		return nil, errors.Wrap(err, "failed to save index")
	}

	return &snapshot, nil
}

// List returns all snapshots in chronological order.
func (m *Manager) List() ([]Snapshot, error) {
	if !m.config.IsEnabled() {
		return nil, ErrBackupDisabled
	}

	if !m.storage.Exists() {
		return []Snapshot{}, nil
	}

	index, err := m.storage.LoadIndex()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load index")
	}

	return index.List(), nil
}

// Get retrieves a snapshot by ID.
func (m *Manager) Get(id string) (*Snapshot, error) {
	if !m.config.IsEnabled() {
		return nil, ErrBackupDisabled
	}

	if !m.storage.Exists() {
		return nil, ErrSnapshotNotFound
	}

	index, err := m.storage.LoadIndex()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load index")
	}

	snapshot, err := index.Get(id)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// determineStorageType determines whether to create a full or patch snapshot.
//
//nolint:unparam // Will be dynamic when delta logic is implemented in Phase 3
func (*Manager) determineStorageType(index *SnapshotIndex) StorageType {
	snapshots := index.List()
	if len(snapshots) == 0 {
		return StorageTypeFull
	}

	// For Phase 1, always create full snapshots
	// Delta logic will be implemented in Phase 3
	return StorageTypeFull
}

// generateChainID generates a chain ID for the snapshot.
//
//nolint:unparam // Will be dynamic when multi-chain logic is implemented
func (*Manager) generateChainID(index *SnapshotIndex) string {
	snapshots := index.List()
	if len(snapshots) == 0 {
		return "chain-1"
	}

	// For Phase 1, use a single chain
	// Multi-chain logic will be implemented when delta is added
	return "chain-1"
}

// getNextSequenceNumber returns the next sequence number for a chain.
func (*Manager) getNextSequenceNumber(index *SnapshotIndex, chainID string) int {
	chain := index.GetChain(chainID)
	if len(chain) == 0 {
		return 1
	}

	maxSeq := 0

	for _, snapshot := range chain {
		if snapshot.SequenceNum > maxSeq {
			maxSeq = snapshot.SequenceNum
		}
	}

	return maxSeq + 1
}

// determineConfigType determines whether a config path is global or project.
func (*Manager) determineConfigType(configPath string) ConfigType {
	// Check if path contains .klaudiush directory (project config)
	// vs ~/.klaudiush (global config)
	// This is a simple heuristic for Phase 1
	if contains(configPath, "/.klaudiush/") {
		return ConfigTypeProject
	}

	return ConfigTypeGlobal
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
