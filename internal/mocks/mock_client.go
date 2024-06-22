package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(key string) (*http.Response, error) {
	args := m.Called(key)
	return args.Get(0).(*http.Response), args.Error(1)
}
