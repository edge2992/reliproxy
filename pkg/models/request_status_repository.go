package models

import "gorm.io/gorm"

type GormRequestStatusRepository struct {
	db *gorm.DB
}

func NewGormRequestStatusRepository(db *gorm.DB) *GormRequestStatusRepository {
	return &GormRequestStatusRepository{db}
}

func (r *GormRequestStatusRepository) GetByID(id string) (*RequestStatus, error) {
	var requestStatus RequestStatus
	err := r.db.First(&requestStatus, "id = ?", id).Error
	return &requestStatus, err
}

func (r *GormRequestStatusRepository) Create(requestStatus *RequestStatus) error {
	return r.db.Create(requestStatus).Error
}
