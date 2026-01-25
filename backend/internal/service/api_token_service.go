package service

import (
	"encoding/json"
	"errors"
	"net"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type APITokenService struct {
	db *gorm.DB
}

func NewAPITokenService(db *gorm.DB) *APITokenService {
	return &APITokenService{db: db}
}

func (s *APITokenService) CreateToken(name string, ipAllowlist []string) (string, *model.APIToken, error) {
	rawToken, tokenHash, err := model.GenerateAPIToken()
	if err != nil {
		return "", nil, err
	}

	ipJSON, err := json.Marshal(ipAllowlist)
	if err != nil {
		return "", nil, err
	}

	token := &model.APIToken{
		UserID:      model.DefaultUserID,
		Name:        name,
		TokenHash:   tokenHash,
		IPAllowlist: datatypes.JSON(ipJSON),
	}

	if err := s.db.Create(token).Error; err != nil {
		return "", nil, err
	}

	return rawToken, token, nil
}

func (s *APITokenService) ListTokens() ([]model.APIToken, error) {
	var tokens []model.APIToken
	if err := s.db.Where("user_id = ?", model.DefaultUserID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *APITokenService) DeleteToken(id uint) error {
	return s.db.Where("user_id = ?", model.DefaultUserID).Delete(&model.APIToken{}, id).Error
}

func (s *APITokenService) ValidateToken(rawToken, clientIP string) (*model.APIToken, error) {
	tokenHash := model.HashToken(rawToken)

	var token model.APIToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}

	if err := validateIP(clientIP, token.IPAllowlist); err != nil {
		return nil, err
	}

	_ = token.UpdateUsage(s.db, clientIP)

	return &token, nil
}

func validateIP(clientIP string, allowlistJSON datatypes.JSON) error {
	var allowlist []string
	if len(allowlistJSON) > 0 {
		if err := json.Unmarshal(allowlistJSON, &allowlist); err != nil {
			return errors.New("invalid ip allowlist")
		}
	}
	var activeRules []string
	for _, r := range allowlist {
		if strings.TrimSpace(r) != "" {
			activeRules = append(activeRules, strings.TrimSpace(r))
		}
	}

	if len(activeRules) == 0 {
		return nil
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return errors.New("invalid client ip")
	}

	for _, rule := range activeRules {
		if strings.Contains(rule, "/") {
			_, ipNet, err := net.ParseCIDR(rule)
			if err != nil {
				continue
			}
			if ipNet.Contains(ip) {
				return nil
			}
		} else {
			if rule == clientIP {
				return nil
			}
		}
	}

	return errors.New("ip not allowed")
}
