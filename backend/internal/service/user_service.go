package service

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetAdmin 获取管理员用户
func (s *UserService) GetAdmin() (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, model.DefaultUserID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// EnsureAdminExists 确保管理员用户存在
func (s *UserService) EnsureAdminExists() error {
	var count int64
	s.db.Model(&model.User{}).Where("id = ?", model.DefaultUserID).Count(&count)
	if count == 0 {
		user := model.User{
			ID: model.DefaultUserID,
		}
		return s.db.Create(&user).Error
	}
	return nil
}

// IsInitialized 检查是否已初始化密码
func (s *UserService) IsInitialized() bool {
	user, err := s.GetAdmin()
	if err != nil {
		return false
	}
	return user.PasswordHash != ""
}

// ValidatePasswordComplexity 验证密码复杂度
func (s *UserService) ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return errors.New("password must contain at least one uppercase letter")
	}
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return errors.New("password must contain at least one lowercase letter")
	}
	if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
		return errors.New("password must contain at least one number")
	}
	return nil
}

// SetupAdmin 设置管理员密码
func (s *UserService) SetupAdmin(password string) error {
	if s.IsInitialized() {
		return errors.New("system already initialized")
	}

	if err := s.ValidatePasswordComplexity(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&model.User{}).Where("id = ?", model.DefaultUserID).Update("password_hash", string(hashedPassword)).Error
}

// VerifyPassword 验证密码
func (s *UserService) VerifyPassword(password string) bool {
	user, err := s.GetAdmin()
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(currentPassword, newPassword string) error {
	if !s.VerifyPassword(currentPassword) {
		return errors.New("current password is incorrect")
	}

	if err := s.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&model.User{}).Where("id = ?", model.DefaultUserID).Update("password_hash", string(hashedPassword)).Error
}
