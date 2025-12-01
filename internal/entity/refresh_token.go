package entity

import "time"

type RefreshToken struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    int64     `gorm:"not null;column:user_id"`
	Token     string    `gorm:"not null;unique"`
	ExpiryAt  time.Time `gorm:"not null;type:timestamp"`
	IsRevoked bool      `gorm:"default:false;not null"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Audit
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func NewRefreshToken(userID int64, token string, expiryAt time.Time, isRevoked bool, user User) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiryAt:  expiryAt,
		IsRevoked: isRevoked,
		User:      user,
	}
}

func (rf *RefreshToken) GetID() int64 {
	return rf.ID
}

func (rf *RefreshToken) GetUserID() int64 {
	return rf.UserID
}

func (rf *RefreshToken) GetToken() string {
	return rf.Token
}

func (rf *RefreshToken) IsExpired() bool {
	return time.Now().After(rf.ExpiryAt)
}

func (rf *RefreshToken) Revoke() {
	rf.IsRevoked = true
}

func (rf *RefreshToken) IsValid() bool {
	return !rf.IsRevoked && !rf.IsExpired()
}
