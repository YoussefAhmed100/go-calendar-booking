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



// Register handles user registration. It creates a new user in the database and
//  generates a JWT token for the user. The function takes a context and a RegisterRequest DTO as input
//  and returns the created User, a JWT token, and an error if any occurred during the process.
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
		"",
	)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// Login handles user login. It verifies the user's credentials, creates a new session in the database,

func (s *AuthService) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (*model.User, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var user model.User

	err := s.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.New("invalid email or password")
	}

	if err != nil {
		return nil, "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	session := model.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: s.jwtService.Expiry(),
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, "", err
	}

	token, err := s.jwtService.GenerateToken(
		user.ID.String(),
		session.ID.String(),
	)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}