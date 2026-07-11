package service

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type UserService struct {
	cfg *config.Config
	db  *gorm.DB
}

var (
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
	ErrAlreadyInitialized       = errors.New("system already initialized")
)

func NewUserService(cfg *config.Config, db *gorm.DB) *UserService {
	return &UserService{cfg: cfg, db: db}
}

func (s *UserService) GetAdmin() (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, model.DefaultUserID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) EnsureAdminExists() error {
	user := model.User{ID: model.DefaultUserID}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user).Error
}

func (s *UserService) IsInitialized() bool {
	user, err := s.GetAdmin()
	if err != nil {
		return false
	}
	return user.PasswordHash != ""
}

func (s *UserService) policy() config.PasswordPolicy {
	if s.cfg == nil {
		return config.PasswordPolicy{MinLength: 8, RequireUpper: true, RequireLower: true, RequireDigit: true}
	}
	return s.cfg.Effective().PasswordPolicy
}

// ValidatePasswordComplexity 按当前 effective 密码策略校验。
func (s *UserService) ValidatePasswordComplexity(password string) error {
	p := s.policy()
	if p.MinLength <= 0 {
		p.MinLength = 8
	}
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.MinLength)
	}
	hasUpper, hasLower, hasDigit, hasSymbol := scanPasswordChars(password)
	if p.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if p.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one number")
	}
	if p.RequireSymbol && !hasSymbol {
		return errors.New("password must contain at least one symbol")
	}
	return nil
}

func scanPasswordChars(password string) (hasUpper, hasLower, hasDigit, hasSymbol bool) {
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>/?`~", r):
			hasSymbol = true
		}
	}
	return
}

func (s *UserService) SetupAdmin(password string) error {
	if err := s.ValidatePasswordComplexity(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		user := model.User{ID: model.DefaultUserID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&user).Error; err != nil {
			return err
		}
		result := tx.Model(&model.User{}).
			Where("id = ? AND COALESCE(password_hash, '') = ''", model.DefaultUserID).
			Update("password_hash", string(hashedPassword))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrAlreadyInitialized
		}
		return nil
	})
}

func (s *UserService) VerifyPassword(password string) bool {
	user, err := s.GetAdmin()
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

func (s *UserService) ChangePassword(currentPassword, newPassword string) error {
	if !s.VerifyPassword(currentPassword) {
		return ErrCurrentPasswordIncorrect
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
