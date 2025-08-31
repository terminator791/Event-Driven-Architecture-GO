package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Exists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// MockProducer is a mock implementation of ProducerInterface
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, key, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	req := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("Exists", req.Email).Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Act
	response, err := service.CreateUser(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Email, response.Email)
	assert.NotEmpty(t, response.ID)
	assert.False(t, response.CreatedAt.IsZero())

	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestUserService_CreateUser_UserAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	req := models.CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("Exists", req.Email).Return(true, nil)

	// Act
	response, err := service.CreateUser(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationFailure(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	testCases := []struct {
		name    string
		request models.CreateUserRequest
		errMsg  string
	}{
		{
			name: "empty email",
			request: models.CreateUserRequest{
				Email:    "",
				Password: "password123",
			},
			errMsg: "email is required",
		},
		{
			name: "empty password",
			request: models.CreateUserRequest{
				Email:    "test@example.com",
				Password: "",
			},
			errMsg: "password is required",
		},
		{
			name: "short password",
			request: models.CreateUserRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			errMsg: "password must be at least 8 characters long",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			response, err := service.CreateUser(context.Background(), tc.request)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	req := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("Exists", req.Email).Return(false, errors.New("database error"))

	// Act
	response, err := service.CreateUser(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to check user existence")

	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_CreateRepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	req := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("Exists", req.Email).Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("create error"))

	// Act
	response, err := service.CreateUser(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_PasswordHashing(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer).(*userService)

	password := "testpassword123"

	// Act
	hashedPassword, err := service.hashPassword(password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Verify the hash can be verified
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestUserService_ValidationLogic(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer).(*userService)

	testCases := []struct {
		name      string
		request   models.CreateUserRequest
		shouldErr bool
	}{
		{
			name: "valid request",
			request: models.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			shouldErr: false,
		},
		{
			name: "invalid email",
			request: models.CreateUserRequest{
				Email:    "",
				Password: "password123",
			},
			shouldErr: true,
		},
		{
			name: "invalid password length",
			request: models.CreateUserRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := service.validateCreateUserRequest(tc.request)

			// Assert
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_EventPublishing_HandlesError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockProducer := new(MockProducer)
	service := NewUserService(mockRepo, mockProducer)

	req := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations - event publishing fails but user creation should still succeed
	mockRepo.On("Exists", req.Email).Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("kafka error"))

	// Act
	response, err := service.CreateUser(context.Background(), req)

	// Assert - should still succeed even if event publishing fails
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Email, response.Email)

	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestUserService_EventData_Validation(t *testing.T) {
	// Test that we can create and serialize events properly
	event := events.NewUserCreated("test-id", "test@example.com", time.Now())

	// Act
	data, err := event.ToJSON()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify we can deserialize it back
	var deserializedEvent events.UserCreated
	err = deserializedEvent.FromJSON(data)
	assert.NoError(t, err)
	assert.Equal(t, event.UserID, deserializedEvent.UserID)
	assert.Equal(t, event.Email, deserializedEvent.Email)
	assert.Equal(t, event.GetEventType(), "user.created")
	assert.NotEmpty(t, event.GetMetadata().EventID)
}