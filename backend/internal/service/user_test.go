package service_test

import (
	"errors"
	"os"
	"testing"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/config"
	"piemdm/pkg/helper/sid"
	jwt2 "piemdm/pkg/jwt"
	"piemdm/pkg/log"
	mock_repository "piemdm/test/mocks/repository"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var (
	logger *log.Logger
	jwtSrv *jwt2.JWT
	sf     *sid.Sid
)

func TestMain(m *testing.M) {
	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); err == nil {
		_ = os.Setenv("APP_CONF", configPath)
	}

	conf := config.NewConfig()
	logger = log.NewLog(conf)
	logger.Info("begin")
	jwtSrv = jwt2.NewJwt(conf)
	sf = sid.NewSid()

	code := m.Run()
	logger.Info("test end")

	os.Exit(code)
}

func TestUserService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	req := &model.User{
		Username: "testuser",
		Password: "password",
		Email:    "test@example.com",
	}

	mockUserRepo.EXPECT().FindByUsername(nil, req.Username).Return(nil, nil)
	mockUserRepo.EXPECT().Create(nil, gomock.Any()).Return(nil)

	err := userService.Create(nil, req)

	assert.NoError(t, err)
}

func TestUserService_Create_UsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	req := &model.User{
		Username: "testuser",
		Password: "password",
		Email:    "test@example.com",
	}

	mockUserRepo.EXPECT().FindByUsername(nil, req.Username).Return(&model.User{}, nil)

	err := userService.Create(nil, req)

	assert.Error(t, err)
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	req := &request.LoginRequest{
		Username: "testuser",
		Password: "password",
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		t.Error("failed to hash password")
	}

	mockUserRepo.EXPECT().FindByUsername(nil, req.Username).Return(&model.User{
		ID:       1,
		Password: string(hashedPassword),
		Email:    "test@example.com",
	}, nil)

	mockUserRepo.EXPECT().GetUserPermissions(gomock.Any(), uint(1)).Return([]string{}, nil)

	token, user, err := userService.Login(nil, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	req := &request.LoginRequest{
		Username: "testuser",
		Password: "password",
	}

	mockUserRepo.EXPECT().FindByUsername(nil, req.Username).Return(nil, errors.New("user not found"))

	_, _, err := userService.Login(nil, req)

	assert.Error(t, err)
}

func TestUserService_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	userId := uint(123)

	mockUserRepo.EXPECT().FindOne(userId).Return(&model.User{
		ID:       userId,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil)

	user, err := userService.Get(userId)

	assert.NoError(t, err)
	assert.Equal(t, userId, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUserService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	userId := uint(123)
	req := &model.User{
		ID:          userId,
		Username:    "testuser",
		DisplayName: "testuser",
		Email:       "test@example.com",
		Password:    "",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	mockUserRepo.EXPECT().FindOne(userId).Return(&model.User{
		ID:       userId,
		Username: "testuser",
		Email:    "old@example.com",
		Password: string(hashedPassword),
	}, nil)
	mockUserRepo.EXPECT().Update(nil, gomock.Any()).Return(nil)

	err := userService.Update(nil, req)

	assert.NoError(t, err)
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	userService := service.NewUserService(logger, sf, jwtSrv, nil, mockUserRepo, mockUserRoleRepo)

	userId := uint(123)
	req := &model.User{
		ID:          userId,
		DisplayName: "testuser",
		Email:       "test@example.com",
	}

	mockUserRepo.EXPECT().FindOne(userId).Return(nil, errors.New("user not found"))

	err := userService.Update(nil, req)

	assert.Error(t, err)
}
