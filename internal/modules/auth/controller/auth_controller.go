package controller

import (
	"net/http"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/modules/auth/dto"
	"calendar-booking/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (controller *AuthController) Register(ctx *gin.Context) {
	var registerRequest dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&registerRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	user, token, err := controller.authService.Register(
		ctx.Request.Context(),
		registerRequest,
	)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (controller *AuthController) Login(ctx *gin.Context) {
	var loginRequest dto.LoginRequest

	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	user, token, err := controller.authService.Login(
		ctx.Request.Context(),
		loginRequest,
	)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (controller *AuthController) Logout(ctx *gin.Context) {
	authenticatedSessionID, exists := commonmiddleware.GetAuthenticatedSessionID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required. Please log in.",
		})
		return
	}

	err := controller.authService.Logout(
		ctx.Request.Context(),
		authenticatedSessionID,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log out. Please try again later.",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "You have been logged out successfully.",
	})
}


func (controller *AuthController) GetCurrentUser(ctx *gin.Context) {
    authenticatedUserID, exists := commonmiddleware.GetAuthenticatedUserID(ctx)
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "error": "Authentication required. Please log in.",
        })
        return
    }

    user, err := controller.authService.GetCurrentUser(
        ctx.Request.Context(),
        authenticatedUserID,
    )
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to retrieve user information.",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "user": gin.H{
            "id":    user.ID,
            "email": user.Email,
        },
    })
}