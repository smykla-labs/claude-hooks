// Package config provides configuration schema types for klaudiush validators.
package config

// BackupConfig contains configuration for the backup system.
type BackupConfig struct {
	// Enabled controls whether the backup system is active.
	// Default: true
	Enabled *bool `json:"enabled,omitempty" koanf:"enabled" toml:"enabled"`

	// AutoBackup controls whether backups are created automatically before config changes.
	// Default: true
	AutoBackup *bool `json:"auto_backup,omitempty" koanf:"auto_backup" toml:"auto_backup"`

	// MaxBackups is the maximum number of backups to keep per config directory.
	// Default: 10
	MaxBackups *int `json:"max_backups,omitempty" koanf:"max_backups" toml:"max_backups"`

	// MaxAge is the maximum age of backups before they are pruned.
	// Default: "720h" (30 days)
	MaxAge Duration `json:"max_age,omitempty" koanf:"max_age" toml:"max_age"`

	// MaxSize is the maximum total size of all backups in bytes.
	// Default: 52428800 (50MB)
	MaxSize *int64 `json:"max_size,omitempty" koanf:"max_size" toml:"max_size"`

	// AsyncBackup controls whether backups run asynchronously.
	// Default: true
	AsyncBackup *bool `json:"async_backup,omitempty" koanf:"async_backup" toml:"async_backup"`

	// Delta contains configuration for delta backup strategy.
	Delta *DeltaConfig `json:"delta,omitempty" koanf:"delta" toml:"delta"`
}

// DeltaConfig contains configuration for delta backup strategy.
type DeltaConfig struct {
	// FullSnapshotInterval is the number of backups between full snapshots.
	// Default: 10
	FullSnapshotInterval *int `json:"full_snapshot_interval,omitempty" koanf:"full_snapshot_interval" toml:"full_snapshot_interval"`

	// FullSnapshotMaxAge is the maximum age before creating a new full snapshot.
	// Default: "168h" (7 days)
	FullSnapshotMaxAge Duration `json:"full_snapshot_max_age,omitempty" koanf:"full_snapshot_max_age" toml:"full_snapshot_max_age"`
}

// IsEnabled returns whether the backup system is enabled.
func (b *BackupConfig) IsEnabled() bool {
	if b == nil || b.Enabled == nil {
		return true
	}

	return *b.Enabled
}

// IsAutoBackupEnabled returns whether automatic backups are enabled.
func (b *BackupConfig) IsAutoBackupEnabled() bool {
	if b == nil || b.AutoBackup == nil {
		return true
	}

	return *b.AutoBackup
}

// IsAsyncBackupEnabled returns whether async backups are enabled.
func (b *BackupConfig) IsAsyncBackupEnabled() bool {
	if b == nil || b.AsyncBackup == nil {
		return true
	}

	return *b.AsyncBackup
}

// GetDelta returns the delta config, creating it if it doesn't exist.
func (b *BackupConfig) GetDelta() *DeltaConfig {
	if b.Delta == nil {
		b.Delta = &DeltaConfig{}
	}

	return b.Delta
}
