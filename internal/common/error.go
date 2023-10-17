package common

import (
	"encoding/json"
	"fmt"
)

type ErrorCodeType uint

const (
	ErrUnableToParseRequestBody  ErrorCodeType = 1000
	ErrUnknownDatabaseError      ErrorCodeType = 1001
	ErrTranslationService        ErrorCodeType = 1002
	ErrInvalidUserID             ErrorCodeType = 1003
	ErrCreateGlucoseRecord       ErrorCodeType = 1004
	ErrInvalidRecordUUID         ErrorCodeType = 1005
	ErrRecordUUIDNotFound        ErrorCodeType = 1006
	ErrGetGlucoseRecord          ErrorCodeType = 1007
	ErrMarkSyncComplete          ErrorCodeType = 1008
	ErrSaveNextChangesToken      ErrorCodeType = 1009
	ErrGetUnsyncedGlucoseRecords ErrorCodeType = 1010
)

type CustomError interface {
	error
	Has(errorCode ErrorCodeType) bool
}

type APIError struct {
	Code    ErrorCodeType `json:"code"`
	Message string        `json:"message"`
}

var ErrorMap = map[ErrorCodeType]string{
	ErrUnableToParseRequestBody:  "invalid request body",
	ErrUnknownDatabaseError:      "database problems",
	ErrTranslationService:        "translation service failed",
	ErrInvalidUserID:             "invalid userId provided in request",
	ErrCreateGlucoseRecord:       "failed to create glucose record",
	ErrInvalidRecordUUID:         "invalid recordUuid provided in request",
	ErrRecordUUIDNotFound:        "glucose record not found",
	ErrGetGlucoseRecord:          "failed to retrieve glucose record",
	ErrMarkSyncComplete:          "failed to mar sync completion",
	ErrSaveNextChangesToken:      "failed to save next changes token",
	ErrGetUnsyncedGlucoseRecords: "faield to retrieve unsynced glucose data",
}

func NewError(errorCode ErrorCodeType) *APIError {
	val, ok := ErrorMap[errorCode]
	if !ok {
		panic(fmt.Errorf("unexpected errorCode-%v", errorCode))
	}

	return &APIError{
		Code:    errorCode,
		Message: val,
	}
}

func (e *APIError) Error() string {
	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (e *APIError) Has(errorCode ErrorCodeType) bool {
	return e.Code == errorCode
}
