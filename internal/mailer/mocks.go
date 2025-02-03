package mailer

import "github.com/stretchr/testify/mock"

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Send(templateFile, username, email string, data any, isSandBox bool) error {
	args := m.Called(templateFile, username, email, data, isSandBox)
	return args.Error(0)
}
