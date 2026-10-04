package tests

import (
	"log"
	"testing"

	"calendar-booking/internal/modules/auth/service"
	"github.com/stretchr/testify/require"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateAndParseToken(t *testing.T) {
	jwtService := service.NewJWTService("test-secret", 1)

	token, err := jwtService.GenerateToken(
		"user-123",
		"session-456",
	)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtService.ParseToken(token)

	require.NoError(t, err)
	require.Equal(t, "user-123", claims.Subject)
	require.Equal(t, "session-456", claims.ID)
}


func TestJWTService_RejectsInvalidAlgorithm(t *testing.T) {
	service := service.NewJWTService("test-secret", 1)

	// log all to show it 



	token := jwt.NewWithClaims(
		jwt.SigningMethodHS384,
		jwt.RegisteredClaims{
			Subject: "user-123",
			ID:      "session-456",
		},
	)



	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	log.Printf("Generated token with HS384: %s", tokenString)

	_, err = service.ParseToken(tokenString)

	require.Error(t, err)
}
