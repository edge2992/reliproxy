package repository

type RequestStatus struct {
	ID     string `gorm:"primary_key"`
	Status string
}

type RequestStatusRepository interface {
	GetByID(id string) (*RequestStatus, error)
	Create(requestStatus *RequestStatus) error
}
