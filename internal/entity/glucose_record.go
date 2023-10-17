package entity

import (
	"time"

	"gorm.io/gorm"
)

type GlucoseLevelType uint8

const (
	MilliMolesPerLiter     GlucoseLevelType = 1
	MilliGramsPerDeciLiter GlucoseLevelType = 2
)

type SpecimenSourceType uint8

const (
	UnknownSource     SpecimenSourceType = 0
	InterstitialFluid SpecimenSourceType = 1
	CapillaryBlood    SpecimenSourceType = 2
	Plasma            SpecimenSourceType = 3
	Serum             SpecimenSourceType = 4
	Tears             SpecimenSourceType = 5
	WholeBlood        SpecimenSourceType = 6
)

type RelationToMealType uint8

const (
	UnknownRelation RelationToMealType = 0
	General         RelationToMealType = 1
	Fasting         RelationToMealType = 2
	BeforeMeal      RelationToMealType = 3
	AfterMeal       RelationToMealType = 4
)

type MealType uint8

const (
	UnknownMeal MealType = 0
	Breakfast   MealType = 1
	Lunch       MealType = 2
	Dinner      MealType = 3
	Snack       MealType = 4
)

// GlucoseRecord -.
type GlucoseRecord struct {
	HealthRecord

	Time           string             `json:"time"`
	ZoneID         string             `json:"zoneId"`
	OffsetID       string             `json:"offsetId"`
	Level          float32            `json:"level"`
	LevelType      GlucoseLevelType   `json:"levelType"`
	SpecimenSource SpecimenSourceType `json:"specimenSource"`
	RelationToMeal RelationToMealType `json:"relationToMeal"`
	Meal           MealType           `json:"meal"`
}

// GlucoseRecordSyncMeta -.
type GlucoseRecordSyncMeta struct {
	gorm.Model
	UserID                        string `json:"userId"`
	AppStorageLastSync            time.Time
	HealthConnectNextChangesToken string
}
