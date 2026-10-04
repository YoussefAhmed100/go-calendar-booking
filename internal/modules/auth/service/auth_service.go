package service

import (
	"context"
	"errors"
	"strings"

	"calendar-booking/internal/modules/auth/dto"
	"calendar-booking/internal/modules/auth/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	jwtService *JWTService
}

func NewAuthService(
	db *gorm.DB,
	jwtService *JWTService,
) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req dto.RegisterRequest,
) (*model.User, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var user model.User

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingUser model.User

		err := tx.
			Where("email = ?", email).
			First(&existingUser).Error

		if err == nil {
			return errors.New("email already exists")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(req.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return err
		}

		user = model.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(passwordHash),
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		session := &model.Session{
			ID:        uuid.New(),
			UserID:    user.ID,
			ExpiresAt: s.jwtService.Expiry(),
		}

		if err := tx.Create(session).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	token, err := s.jwtService.GenerateToken(
		user.ID.String(),
		// هنحتاج session ID هنا
		"",
	)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}