package service

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type APITokenService struct {
	db *gorm.DB
}

var ErrInvalidTokenType = errors.New("invalid token type")

func NewAPITokenService(db *gorm.DB) *APITokenService {
	return &APITokenService{db: db}
}

func (s *APITokenService) CreateToken(name string, ipAllowlist []string, tokenType string) (string, *model.APIToken, error) {
	rawToken, tokenHash, err := model.GenerateAPIToken()
	if err != nil {
		return "", nil, err
	}

	if tokenType == "" {
		tokenType = model.TokenTypeFull
	}
	switch tokenType {
	case model.TokenTypeFull, model.TokenTypeUploadList, model.TokenTypeListOnly:
	default:
		return "", nil, ErrInvalidTokenType
	}

	ipJSON, err := json.Marshal(ipAllowlist)
	if err != nil {
		return "", nil, err
	}

	token := &model.APIToken{
		UserID:      model.DefaultUserID,
		Name:        name,
		TokenType:   tokenType,
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

	if token.TokenType == "" {
		token.TokenType = model.TokenTypeFull
	}

	_ = token.UpdateUsage(s.db, clientIP)

	return &token, nil
}

func (s *APITokenService) GetTokenByID(id uint) (*model.APIToken, error) {
	var token model.APIToken
	if err := s.db.Where("user_id = ?", model.DefaultUserID).First(&token, id).Error; err != nil {
		return nil, err
	}
	if token.TokenType == "" {
		token.TokenType = model.TokenTypeFull
	}
	return &token, nil
}

func (s *APITokenService) RecordLog(log *model.APITokenLog) error {
	if log == nil {
		return nil
	}
	return s.db.Create(log).Error
}

func (s *APITokenService) ListLogs(page, pageSize int) ([]model.APITokenLog, int64, error) {
	var logs []model.APITokenLog
	var total int64
	query := s.db.Model(&model.APITokenLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *APITokenService) CleanupLogsBefore(cutoff time.Time) (int64, error) {
	result := s.db.Where("created_at < ?", cutoff).Delete(&model.APITokenLog{})
	return result.RowsAffected, result.Error
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
