package model

import (
	"encoding/base64"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

const DefaultUserID = 1

type PasskeyCredential struct {
	ID              uint64 `gorm:"primaryKey"`
	CredentialID    string `gorm:"size:512;uniqueIndex;not null"`
	PublicKey       []byte `gorm:"type:bytea;not null"`
	AttestationType string `gorm:"size:64"`
	AAGUID          []byte `gorm:"column:aaguid;type:bytea"`
	SignCount       uint32
	UserID          uint64 `gorm:"not null"`
	UserAgent       string `gorm:"type:text"`
	IPAddress       string `gorm:"size:45"`
	DeviceName      string `gorm:"size:255"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (c *PasskeyCredential) ToWebAuthnCredential() webauthn.Credential {
	credentialID, _ := base64.RawURLEncoding.DecodeString(c.CredentialID)
	
	return webauthn.Credential{
		ID:              credentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Authenticator: webauthn.Authenticator{
			AAGUID:    c.AAGUID,
			SignCount: c.SignCount,
		},
	}
}

func FromWebAuthnCredential(cred webauthn.Credential, userAgent, ipAddress, deviceName string) *PasskeyCredential {
	credentialID := base64.RawURLEncoding.EncodeToString(cred.ID)
	
	return &PasskeyCredential{
		CredentialID:    credentialID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		AAGUID:          cred.Authenticator.AAGUID,
		SignCount:       cred.Authenticator.SignCount,
		UserAgent:       userAgent,
		IPAddress:       ipAddress,
		DeviceName:      deviceName,
	}
}

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	PasswordHash string `gorm:"size:255"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Credentials  []PasskeyCredential `gorm:"foreignKey:UserID"`
	APITokens    []APIToken          `gorm:"foreignKey:UserID"`
}

func (u *User) WebAuthnID() []byte {
	return []byte{byte(u.ID)}
}


func (u *User) WebAuthnName() string {
	return "admin"
}

func (u *User) WebAuthnDisplayName() string {
	return "Administrator"
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = c.ToWebAuthnCredential()
	}
	return creds
}

func (u *User) WebAuthnIcon() string {
	return ""
}
