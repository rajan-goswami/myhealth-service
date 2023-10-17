// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"myhealth-service/internal/common"
	"myhealth-service/internal/entity"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// GlucoseDataRepo -.
	GlucoseDataRepo interface {
		Create(*entity.GlucoseRecord) (uuid.UUID, error)
		CreateGlucoseRecordSyncMeta(*entity.GlucoseRecordSyncMeta) (uint, error)
		IsDeviceRecordExists(userID string, deviceRecordID string) bool
		GetByRecordUUID(userID string, recordUUID uuid.UUID) (entity.GlucoseRecord, error)
		GetRecordsSinceLastSync(userID string) ([]entity.GlucoseRecord, error)
		MarkSyncComplete(userID string) error
		SaveNextChangesToken(userID string, token string) error
	}

	// GlucoseTracking -.
	GlucoseTracking interface {
		// CreateRecord creates a new record with glucose measurement details of a user.
		CreateRecord(ctx context.Context, gr entity.GlucoseRecord) (uuid.UUID, error)

		// GetNewOrUpdatedRecords returns new or updated glucose records since the last sync.
		GetNewOrUpdatedRecords(ctx context.Context, userID string) ([]entity.GlucoseRecord, error)

		GetRecord(ctx context.Context, userID string, recordUuid uuid.UUID) (entity.GlucoseRecord, common.CustomError)

		MarkSyncComplete(ctx context.Context, userID string) common.CustomError

		SaveNextChangesToken(ctx context.Context, userID string, nextChangesToken string) common.CustomError
	}
)
