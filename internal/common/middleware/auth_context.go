package middleware

import "github.com/gin-gonic/gin"

const (
    AuthenticatedUserIDKey = "authenticated_user_id"
    AuthenticatedSessionIDKey = "authenticated_session_id"
)

func GetAuthenticatedUserID(c *gin.Context) (string, bool) {
    value, exists := c.Get(AuthenticatedUserIDKey)
    if !exists {
        return "", false
    }

    userID, ok := value.(string)
    return userID, ok
}

func GetAuthenticatedSessionID(c *gin.Context) (string, bool) {
    value, exists := c.Get(AuthenticatedSessionIDKey)
    if !exists {
        return "", false
    }

    sessionID, ok := value.(string)
    return sessionID, ok
}