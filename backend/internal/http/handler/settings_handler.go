package handler

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type SettingsHandler struct {
	db       *gorm.DB
	settings *service.SettingsService
	cfg      *config.Config
	log      *logger.Logger
}

func NewSettingsHandler(cfg *config.Config, db *gorm.DB, settings *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		db:       db,
		settings: settings,
		cfg:      cfg,
		log:      logger.Register("settings-handler"),
	}
}

type settingsResponse struct {
	Schema          []service.FieldSchema `json:"schema"`
	Values          []service.FieldValue  `json:"values"`
	AllowWebModify  bool                  `json:"allow_web_modify"`
	BootstrapNotice string                `json:"bootstrap_notice,omitempty"`
}

func (h *SettingsHandler) Get(c *gin.Context) {
	values, err := h.settings.CurrentValues()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "settings_load_failed", "failed to load settings")
		return
	}
	resp := settingsResponse{
		Schema:         h.settings.Schema(),
		Values:         values,
		AllowWebModify: h.settings.AllowsWebModify(),
	}
	if !resp.AllowWebModify {
		resp.BootstrapNotice = "ANZUIMG_ALLOW_WEB_CONFIG=false; modifications must be done via .env or DB"
	}
	c.JSON(http.StatusOK, resp)
}

type patchSettingsRequest struct {
	Values map[string]string `json:"values" binding:"required"`
}

func (h *SettingsHandler) Patch(c *gin.Context) {
	if !h.settings.AllowsWebModify() {
		response.WriteErrorCode(c, http.StatusForbidden, "web_config_disabled", "web config modification is disabled")
		return
	}
	var req patchSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Values) == 0 {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_settings_request", "invalid request")
		return
	}

	if err := h.guardSelfLockout(c, req.Values); err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "self_lockout_blocked", err.Error())
		return
	}

	if err := h.settings.Set(req.Values); err != nil {
		if err == service.ErrWebConfigDisabled {
			response.WriteErrorCode(c, http.StatusForbidden, "web_config_disabled", err.Error())
			return
		}
		h.recordSecurityEvent(c, "warning", "settings_update_failed", err.Error())
		response.WriteErrorCode(c, http.StatusBadRequest, "settings_update_failed", err.Error())
		return
	}
	h.recordSecurityEvent(c, "info", "settings_updated", "system settings updated via web")
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

type resetSettingsRequest struct {
	Keys []string `json:"keys" binding:"required"`
}

func (h *SettingsHandler) Reset(c *gin.Context) {
	if !h.settings.AllowsWebModify() {
		response.WriteErrorCode(c, http.StatusForbidden, "web_config_disabled", "web config modification is disabled")
		return
	}
	var req resetSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Keys) == 0 {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_settings_request", "invalid request")
		return
	}
	if err := h.settings.Reset(req.Keys); err != nil {
		h.recordSecurityEvent(c, "warning", "settings_reset_failed", err.Error())
		response.WriteErrorCode(c, http.StatusBadRequest, "settings_reset_failed", err.Error())
		return
	}
	h.recordSecurityEvent(c, "info", "settings_reset", "system settings reset to defaults")
	c.JSON(http.StatusOK, gin.H{"message": "settings reset"})
}

// guardSelfLockout 防止管理员误把自己锁在外面: 如果本次提交涉及黑/白名单,
// 应用变更后管理员当前 IP 必须仍能访问。
func (h *SettingsHandler) guardSelfLockout(c *gin.Context, values map[string]string) error {
	bl, blChanged := values["IP_BLACKLIST"]
	al, alChanged := values["ADMIN_IP_ALLOWLIST"]
	if !blChanged && !alChanged {
		return nil
	}
	myIP := middleware.ClientIP(c)
	if myIP == "" {
		return nil
	}
	if blChanged {
		newList := model.ParseConfigStringList(bl)
		if matched, _ := matchIPInList(myIP, newList); matched {
			return errSelfLockout("you are about to add your own IP to IP_BLACKLIST")
		}
	}
	if alChanged {
		newList := model.ParseConfigStringList(al)
		if len(newList) > 0 {
			if matched, _ := matchIPInList(myIP, newList); !matched {
				return errSelfLockout("ADMIN_IP_ALLOWLIST does not include your current IP")
			}
		}
	}
	return nil
}

func errSelfLockout(msg string) error { return &selfLockoutError{msg: msg} }

type selfLockoutError struct{ msg string }

func (e *selfLockoutError) Error() string { return e.msg }

func (h *SettingsHandler) recordSecurityEvent(c *gin.Context, level, action, message string) {
	username := "admin"
	event := &model.SecurityEventLog{
		Category:  "config",
		Level:     level,
		Action:    action,
		Message:   message,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: middleware.ClientIP(c),
		Username:  username,
	}
	if err := h.db.Create(event).Error; err != nil {
		h.log.Ctx(c.Request.Context()).Warnf("record security event failed: %v", err)
	}
}

// matchIPInList 支持 IP 与 CIDR 形式的列表匹配。
func matchIPInList(ip string, list []string) (bool, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" || len(list) == 0 {
		return false, nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false, err
	}
	for _, entry := range list {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			pfx, err := netip.ParsePrefix(entry)
			if err != nil {
				continue
			}
			if pfx.Contains(addr) {
				return true, nil
			}
		} else {
			other, err := netip.ParseAddr(entry)
			if err != nil {
				continue
			}
			if other == addr {
				return true, nil
			}
		}
	}
	return false, nil
}
