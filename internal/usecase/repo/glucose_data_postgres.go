package repo

import (
	"errors"
	"myhealth-service/internal/entity"
	"myhealth-service/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUnknownDatabaseError     = errors.New("unknown database error")
	ErrGlucoseRecordExists      = errors.New("glucose record with this id already exists")
	ErrUserIDNotFound           = errors.New("this user id does not exists")
	ErrGlucoseRecordNotFound    = errors.New("glucose record with this id does not exists")
	ErrGlucoseRecordMissingData = errors.New("glucose record is missing required data")
	ErrGlucoseRecordInvalidData = errors.New("invalid data provided for a glucose measurement record")
)

// GlucoseDataRepo -.
type GlucoseDataRepo struct {
	db     *gorm.DB
	logger logger.Interface
}

// NewGlucoseDataRepo creates a new instance of glucose data repository.
func NewGlucoseDataRepo(db *gorm.DB, l logger.Interface) *GlucoseDataRepo {
	return &GlucoseDataRepo{db, l}
}

// Create creates a new database record for a glucose measurement record.

// It does not check if record with deviceRecordId exists or not. This check is left to the
// device side application.
func (r *GlucoseDataRepo) Create(gr *entity.GlucoseRecord) (uuid.UUID, error) {
	tnx := r.db.Create(gr)

	if tnx.Error != nil {
		r.logger.Error(tnx.Error, "[Create] error in creating record")
		if errors.Is(tnx.Error, gorm.ErrDuplicatedKey) {
			return uuid.UUID{}, ErrGlucoseRecordExists
		} else if errors.Is(tnx.Error, gorm.ErrModelValueRequired) {
			return uuid.UUID{}, ErrGlucoseRecordMissingData
		} else if errors.Is(tnx.Error, gorm.ErrInvalidData) {
			return uuid.UUID{}, ErrGlucoseRecordInvalidData
		} else {
			return uuid.UUID{}, ErrUnknownDatabaseError
		}
	}
	return gr.UUID, nil
}

func (r *GlucoseDataRepo) CreateGlucoseRecordSyncMeta(gr *entity.GlucoseRecordSyncMeta) (uint, error) {
	tnx := r.db.Create(gr)

	if tnx.Error != nil {
		r.logger.Error(tnx.Error, "[CreateGlucoseRecordSyncMeta] error in creating sync meta")
		return 0, ErrUnknownDatabaseError
	}
	return 0, nil
}

// GetRecordsSinceLastSync returns a glucose records which are created/updated since last device sync
func (r *GlucoseDataRepo) GetRecordsSinceLastSync(userID string) ([]entity.GlucoseRecord, error) {
	var records []entity.GlucoseRecord
	err := r.db.Transaction(func(tx *gorm.DB) error {
		readAll := false
		var meta entity.GlucoseRecordSyncMeta
		err := tx.Select("*").
			Model(&entity.GlucoseRecordSyncMeta{}).
			Where("user_id = ?", userID).
			First(&meta).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// For the first time sync, we should return all records
				readAll = true
			} else {
				return err
			}
		}

		if readAll {
			r.db.Select("*").
				Model(&entity.GlucoseRecord{}).
				Where("user_id = ?", userID).
				Order("updated_at DESC").
				Find(&records)
		} else {
			r.db.Select("*").
				Model(&entity.GlucoseRecord{}).
				Where("user_id = ?", userID).
				Where("updated_at > ?", meta.AppStorageLastSync).
				Order("updated_at DESC").
				Find(&records)
		}

		return nil
	})

	if err != nil {
		// handle database errors
		r.logger.Error(err, "[GetRecordsSinceLastSync] database error occurred")
		return nil, ErrUnknownDatabaseError

	}

	return records, nil
}

// GetByRecordUUID returns a glucose record matching with a
// userID and recordUUID.
func (r *GlucoseDataRepo) GetByRecordUUID(userID string, recordUUID uuid.UUID) (entity.GlucoseRecord, error) {
	var gr entity.GlucoseRecord
	tnx := r.db.Select("*").
		Model(&entity.GlucoseRecord{}).
		Where("user_id = ?", userID).
		Where("uuid = ?", recordUUID).
		Take(&gr)

	if tnx.Error != nil {
		// handle database errors
		if errors.Is(tnx.Error, gorm.ErrRecordNotFound) {
			return entity.GlucoseRecord{}, ErrGlucoseRecordNotFound
		} else {
			r.logger.Error(tnx.Error, "[GetByRecordUUID] database error occurred")
			return entity.GlucoseRecord{}, ErrUnknownDatabaseError
		}
	}

	return gr, nil
}

// IsDeviceRecordExists checks if a glucose record exists with a matching
// userID and deviceRecordID.
func (r *GlucoseDataRepo) IsDeviceRecordExists(userID string, deviceRecordID string) bool {
	var gr entity.GlucoseRecord
	tnx := r.db.Select("*").
		Model(&entity.GlucoseRecord{}).
		Where("user_id = ?", userID).
		Where("device_record_id = ?", deviceRecordID).
		Take(&gr)

	if tnx.Error != nil {
		// handle database errors
		if errors.Is(tnx.Error, gorm.ErrRecordNotFound) {
			return false
		} else {
			r.logger.Error(tnx.Error, "[IsDeviceRecordExists] database error occurred")
			return false
		}
	}

	return true
}

// MarkSyncComplete sets last sync time to current time for a user's device sync operation.
func (r *GlucoseDataRepo) MarkSyncComplete(userID string) error {
	tnx := r.db.Select("*").
		Model(&entity.GlucoseRecordSyncMeta{}).
		Where("user_id = ?", userID).
		Update("app_storage_last_sync", time.Now())

	if tnx.Error != nil {
		r.logger.Error(tnx.Error, "[MarkSyncComplete] database error occurred")
		return ErrUnknownDatabaseError
	}
	if tnx.RowsAffected == 0 {
		return ErrUserIDNotFound
	}

	return nil
}

// SaveNextChangesToken sets next changes token of a user's data sync.
func (r *GlucoseDataRepo) SaveNextChangesToken(userID string, changesToken string) error {
	tnx := r.db.Select("*").
		Model(&entity.GlucoseRecordSyncMeta{}).
		Where("user_id = ?", userID).
		UpdateColumn("health_connect_next_changes_token", changesToken)

	if tnx.Error != nil {
		r.logger.Error(tnx.Error, "[SaveNextChangesToken] database error occurred")
		return ErrUnknownDatabaseError
	}
	if tnx.RowsAffected == 0 {
		return ErrUserIDNotFound
	}

	return nil
}
