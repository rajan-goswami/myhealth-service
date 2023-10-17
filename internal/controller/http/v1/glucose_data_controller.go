package v1

import (
	"myhealth-service/internal/common"
	"myhealth-service/internal/entity"
	"myhealth-service/internal/usecase"
	"myhealth-service/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type glocoseDataController struct {
	g usecase.GlucoseTracking
	l logger.Interface
}

func newGlucoseDataController(handler *gin.RouterGroup, g usecase.GlucoseTracking, l logger.Interface) {
	r := &glocoseDataController{g, l}

	h := handler.Group("/glucose-records")
	{
		h.GET("/unsynced", r.getNewOrUpdatedRecords)
		h.POST("/sync-complete", r.postSyncComplete)
		h.POST("/next-changes-token", r.postNextChangesToken)
		h.POST("/", r.createGlucoseRecord)
		h.GET("/:recordUuid", r.getGlucoseRecord)
	}
}

type createGlucoseRecordRequest struct {
	DeviceRecordID string  `json:"deviceRecordId" binding:"min=5"`
	Time           string  `json:"time" binding:"required"`
	ZoneID         string  `json:"zoneId" binding:"required"`
	OffsetID       string  `jsom:"offsetId" binding:"required"`
	Level          float32 `json:"level" binding:"required"`
	LevelType      uint8   `json:"levelType" binding:"required,min=1,max=2"`
	SpecimenSource uint8   `json:"specimenSource" binding:"required,min=0,max=6"`
	Meal           uint8   `json:"meal" binding:"required,min=0,max=4"`
	RelationToMeal uint8   `json:"relationToMeal" binding:"required,min=0,max=4"`
}

type createGlucoseRecordResponse struct {
	RecordID string `json:"recordId"`
}

// @Summary		CreateGlucoseRecord
// @Description	Post a new glucose level measurement
// @ID				create-glucose-record
// @Tags			glucose
// @Accept			json
// @Produce		json
// @Param			request	body		createGlucoseRecordRequest	true	"Post Glucose Record"
// @Param			id		path		string						true	"User ID"
// @Success		200		{object}	createGlucoseRecordResponse
// @Failure		400		{object}	common.APIError
// @Failure		500		{object}	common.APIError
// @Router			/users/:id/glucose-records [post]
func (r *glocoseDataController) createGlucoseRecord(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidUserID))
		return
	}

	var request createGlucoseRecordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "[createGlucoseRecord] error in parsing request")
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrUnableToParseRequestBody))

		return
	}

	recordUUID, err := r.g.CreateRecord(
		c.Request.Context(),
		entity.GlucoseRecord{
			HealthRecord: entity.HealthRecord{
				DeviceRecordID: request.DeviceRecordID,
				UserID:         userID,
			},
			Time:           request.Time,
			ZoneID:         request.ZoneID,
			OffsetID:       request.OffsetID,
			Level:          request.Level,
			LevelType:      entity.GlucoseLevelType(request.LevelType),
			SpecimenSource: entity.SpecimenSourceType(request.SpecimenSource),
			RelationToMeal: entity.RelationToMealType(request.RelationToMeal),
			Meal:           entity.MealType(request.Meal),
		},
	)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, createGlucoseRecordResponse{
		RecordID: recordUUID.String(),
	})
}

type glucoseRecordsResponse struct {
	Records []entity.GlucoseRecord `json:"records"`
}

// @Summary		Get glucose records which are new/updated since last sync time.
// @Description	Get glucose records which are new/updated since last sync time.
// @ID				getNewOrUpdatedRecords
// @Tags			glucose
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	glucoseRecordsResponse
// @Failure		400	{object}	common.APIError
// @Failure		500	{object}	common.APIError
// @Router			/users/:id/glucose-records/unsynced [get]
func (r *glocoseDataController) getNewOrUpdatedRecords(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidUserID))
		return
	}

	records, err := r.g.GetNewOrUpdatedRecords(c.Request.Context(), userID)
	if err != nil {
		r.l.Error(err, "GetNewOrUpdatedRecords error")
		ErrorResponse(c, http.StatusInternalServerError, common.NewError(common.ErrUnknownDatabaseError))

		return
	}

	c.JSON(http.StatusOK, glucoseRecordsResponse{records})
}

// @Summary		Marks sync as completed for a user's glucose records data on device.
// @Description	Marks sync as completed for a user's glucose records data on device.
// @ID				postSyncComplete
// @Tags			glucose
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200
// @Failure		400	{object}	common.APIError
// @Failure		500	{object}	common.APIError
// @Router			/users/:id/glucose-records/sync-complete [post]
func (r *glocoseDataController) postSyncComplete(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidUserID))
		return
	}

	customErr := r.g.MarkSyncComplete(c.Request.Context(), userID)
	if customErr != nil {
		ErrorResponse(c, http.StatusInternalServerError, customErr)
		return
	}

	c.Status(http.StatusOK)
}

type saveNextChangesTokenRequest struct {
	NextChangesToken string `json:"deviceRecordId" binding:"required,min=5"`
}

// @Summary		Stores device's next changes token of a user's glucose records.
// @Description	Stores device's next changes token of a user's glucose records.
// @ID				postNextChangesToken
// @Tags			glucose
// @Accept			json
// @Produce		json
// @Param			id		path	string						true	"User ID"
// @Param			request	body	saveNextChangesTokenRequest	true	"next changes token"
// @Success		200
// @Failure		400	{object}	common.APIError
// @Failure		500	{object}	common.APIError
// @Router			/users/:id/glucose-records/next-changes-token [post]
func (r *glocoseDataController) postNextChangesToken(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidUserID))
		return
	}

	var request saveNextChangesTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "[postNextChangesToken] error in parsing request")
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrUnableToParseRequestBody))
		return
	}

	customErr := r.g.SaveNextChangesToken(c.Request.Context(), userID, request.NextChangesToken)
	if customErr != nil {
		ErrorResponse(c, http.StatusInternalServerError, customErr)
		return
	}

	c.JSON(http.StatusOK, nil)
}

type glucoseRecordResponse struct {
	entity.GlucoseRecord
}

// @Summary		Get Glucose Measurement Details.
// @Description	Get a glucose measurement record of a user.
// @ID				getGlucoseRecord
// @Tags			glucose
// @Produce		json
// @Param			id			path		string	true	"User ID"
// @Param			recordUuid	path		string	true	"Glucose record UUID"
// @Success		200			{object}	glucoseRecordsResponse
// @Failure		400			{object}	common.APIError
// @Failure		404			{object}	common.APIError
// @Failure		500			{object}	common.APIError
// @Router			/users/:id/glucose-records/:recordUuid [get]
func (r *glocoseDataController) getGlucoseRecord(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidUserID))
		return
	}

	recordUUID, err := uuid.Parse(c.Param("recordUuid"))
	if err != nil {
		r.l.Error(err, "[getGlucoseRecord]: error in parsing recordUuid-%v", c.Param("recordUuid"))
		ErrorResponse(c, http.StatusBadRequest, common.NewError(common.ErrInvalidRecordUUID))
		return
	}

	record, customErr := r.g.GetRecord(c.Request.Context(), userID, recordUUID)
	if customErr != nil {
		if customErr.Has(common.ErrRecordUUIDNotFound) {
			ErrorResponse(c, http.StatusNotFound, customErr)
		} else {
			ErrorResponse(c, http.StatusInternalServerError, customErr)
		}
		return
	}

	c.JSON(http.StatusOK, glucoseRecordResponse{record})
}
