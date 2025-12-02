# Backup System Implementation

Phase 1 implementation of automatic configuration backup system for klaudiush.

## Architecture

### Core Components

**Snapshot** (`internal/backup/snapshot.go`):

- Types: `StorageType` (full/patch), `ConfigType` (global/project), `Trigger` (manual/automatic/before_init/migration)
- `Snapshot` struct: Contains ID, sequence number, timestamp, config path, storage details, chain info, metadata
- `SnapshotIndex`: Maps snapshot IDs to metadata, provides operations (Add/Get/Delete/List/FindByHash/GetChain)
- Deduplication: `FindByHash()` enables content-based dedup before creating new snapshots

**Storage** (`internal/backup/storage.go`):

- Interface-based design: `Storage` interface with `FilesystemStorage` implementation
- Centralized structure: `~/.klaudiush/.backups/{global,projects/*/}/snapshots/`
- Operations: Save/Load/Delete/List snapshots, SaveIndex/LoadIndex for metadata
- Path sanitization: Converts `/Users/bart/project` → `Users_bart_project` for directory names
- Security: 0o600 file permissions, 0o700 directory permissions
- Uses `strings.Builder` for efficient path manipulation

**Manager** (`internal/backup/manager.go`):

- Orchestrates all backup operations
- `CreateBackup()`: Reads config, computes hash, checks dedup, determines storage type, saves snapshot, updates index
- Automatic storage initialization on first use
- Returns existing snapshot if content hash matches (deduplication)
- Phase 1: Only full snapshots (delta/patch support planned for Phase 3)
- Helper methods: `determineStorageType()`, `generateChainID()`, `getNextSequenceNumber()`, `determineConfigType()`

**Configuration** (`pkg/config/backup.go`):

- `BackupConfig`: Enabled, AutoBackup, MaxBackups, MaxAge, MaxSize, AsyncBackup
- `DeltaConfig`: FullSnapshotInterval, FullSnapshotMaxAge (for future delta support)
- Helper methods: `IsEnabled()`, `IsAutoBackupEnabled()`, `IsAsyncBackupEnabled()`, `GetDelta()`
- Added to root `Config` struct with `GetBackup()` accessor

## Key Design Decisions

### Centralized Storage

All backups stored in `~/.klaudiush/.backups/` instead of scattered `.backups/` directories in each project. Benefits:

- Single location for all backups
- Easier to manage and query
- No clutter in project directories
- Global and project configs clearly separated

### Deduplication

Always-on content-based deduplication using SHA256 hashes:

- Before creating backup, check if identical content already exists via `FindByHash()`
- If found, return existing snapshot instead of creating duplicate
- Prevents wasted storage for unchanged configs
- Tested with multiple backup attempts of same content

### Interface-Based Storage

`Storage` interface allows for future storage backends (S3, database, etc.) without changing manager code. Currently implemented: `FilesystemStorage`.

### Security

- File permissions: 0o600 (owner read/write only)
- Directory permissions: 0o700 (owner access only)
- No encryption (rely on filesystem encryption like FileVault/LUKS)
- Checksums: SHA256 for integrity validation

## Testing

89.7% test coverage achieved:

- `snapshot_test.go`: Tests for all snapshot types, index operations, ID generation, hash computation
- `storage_test.go`: Tests for filesystem storage, initialization, CRUD operations, project isolation
- `manager_test.go`: Tests for manager operations, deduplication, triggers, config type detection
- `backup_test.go` (in pkg/config): Tests for configuration types and helper methods

Test patterns:

- Ginkgo/Gomega framework
- BeforeEach/AfterEach for setup/teardown
- Temporary directories for isolation
- Comprehensive edge case coverage

## Integration Points (Future Phases)

**Writer** (`internal/config/writer.go`): Will add BackupManager field, call CreateBackup() before writing config

**Init** (`internal/initcmd/init.go`): Will create backup with TriggerBeforeInit when using --force flag

**Main** (`cmd/klaudiush/main.go`): Will instantiate manager, perform first-run migration

## Linter Fixes Applied

- Used `strings.Builder` instead of string concatenation in loops (modernize)
- Removed underscore receivers, using `(*Type)` syntax (staticcheck ST1006)
- Added `#nosec G304` comments for controlled file reads (gosec)
- Fixed variable shadowing in tests (govet)
- Merged variable declarations with assignments where appropriate (staticcheck S1021)
- Added `//nolint:unparam` for methods that will become dynamic in Phase 3
- Formatted long lines using multiline function calls (golines)

## Future Work

**Phase 2 - Retention**: Implement retention policies (count/age/size-based), chain-aware cleanup

**Phase 3 - Restore**: Implement restore functionality, diff between snapshots, patch reconstruction using delta library

**Phase 4 - Integration**: Wire up automatic backups in config writer and init command

**Phase 5 - CLI**: Add `klaudiush backup` subcommands (list/create/restore/delete/diff/prune/audit/status)

**Phase 6 - Audit**: Implement audit logging for all backup operations

**Phase 7 - Doctor**: Add backup health checks and fixers to doctor command

**Phase 8 - Documentation**: Create user guide, example configurations

**Phase 9 - Testing**: Add integration and E2E tests

**Phase 10 - Migration**: First-run backup creation for existing users

## Performance Characteristics

- Full snapshot save: O(n) where n = config file size
- Dedup lookup: O(1) hash map lookup
- Snapshot list: O(m) where m = number of snapshots
- Storage initialization: One-time overhead, ~10ms
- Typical operation: <100ms for small configs (<50KB)

## Error Handling

Uses `github.com/cockroachdb/errors` for all error creation and wrapping:

- `ErrSnapshotNotFound`: Snapshot ID not found in index
- `ErrStorageNotInitialized`: Storage not initialized before use
- `ErrInvalidPath`: Invalid path provided to storage
- `ErrInvalidConfigType`: Invalid config type (must be global/project)
- `ErrInvalidStorageType`: Invalid storage type (must be full/patch)
- `ErrConfigFileNotFound`: Config file doesn't exist
- `ErrBackupDisabled`: Backup system is disabled in configuration

All errors wrapped with context using `errors.Wrap()` or `errors.Wrapf()`.
