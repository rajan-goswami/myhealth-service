package usecase

import (
	"context"
	"myhealth-service/internal/common"
	"myhealth-service/internal/entity"
	"myhealth-service/internal/usecase/repo"
	"time"

	"github.com/google/uuid"
)

// GlucoseTrackingUseCase -.
type GlucoseTrackingUseCase struct {
	repo GlucoseDataRepo
}

// NewGlucoseTracking -.
func NewGlucoseTracking(r GlucoseDataRepo) *GlucoseTrackingUseCase {
	return &GlucoseTrackingUseCase{
		repo: r,
	}
}

func (u *GlucoseTrackingUseCase) CreateRecord(ctx context.Context, gr entity.GlucoseRecord) (uuid.UUID, error) {
	recordUUID, err := u.repo.Create(&gr)
	if err != nil {
		return uuid.UUID{}, common.NewError(common.ErrCreateGlucoseRecord)
	}
	return recordUUID, nil
}

func (u *GlucoseTrackingUseCase) GetNewOrUpdatedRecords(ctx context.Context, userID string) ([]entity.GlucoseRecord, error) {
	records, err := u.repo.GetRecordsSinceLastSync(userID)
	if err != nil {
		return []entity.GlucoseRecord{}, common.NewError(common.ErrGetUnsyncedGlucoseRecords)
	}
	return records, nil
}

func (u *GlucoseTrackingUseCase) GetRecord(ctx context.Context, userID string, recordUUID uuid.UUID) (entity.GlucoseRecord, common.CustomError) {
	record, err := u.repo.GetByRecordUUID(userID, recordUUID)
	if err != nil {
		if err == repo.ErrGlucoseRecordNotFound {
			return entity.GlucoseRecord{}, common.NewError(common.ErrRecordUUIDNotFound)
		}
		return entity.GlucoseRecord{}, common.NewError(common.ErrGetGlucoseRecord)
	}
	return record, nil
}

func (u *GlucoseTrackingUseCase) MarkSyncComplete(ctx context.Context, userID string) common.CustomError {
	err := u.repo.MarkSyncComplete(userID)
	if err != nil {
		if err == repo.ErrUserIDNotFound {
			// create first time sync metadata
			_, err = u.repo.CreateGlucoseRecordSyncMeta(
				&entity.GlucoseRecordSyncMeta{
					UserID:                        userID,
					AppStorageLastSync:            time.Now(),
					HealthConnectNextChangesToken: "",
				},
			)
			if err == nil {
				return nil
			}
		}
		return common.NewError(common.ErrMarkSyncComplete)
	}
	return nil
}

func (u *GlucoseTrackingUseCase) SaveNextChangesToken(ctx context.Context, userID string, nextChangesToken string) common.CustomError {
	err := u.repo.SaveNextChangesToken(userID, nextChangesToken)
	if err != nil {
		if err == repo.ErrUserIDNotFound {
			// create first time sync metadata
			_, err = u.repo.CreateGlucoseRecordSyncMeta(
				&entity.GlucoseRecordSyncMeta{
					UserID:                        userID,
					AppStorageLastSync:            time.Now(),
					HealthConnectNextChangesToken: "",
				},
			)
			if err == nil {
				return nil
			}
		}
		return common.NewError(common.ErrSaveNextChangesToken)
	}
	return nil
}
