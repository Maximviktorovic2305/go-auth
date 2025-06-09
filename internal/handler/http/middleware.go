package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"server/internal/models"
	"server/internal/service"
	"strings"
	"server/internal/config"
)

type Middleware struct {
	tokenService *service.TokenService
	userService  *service.UserService
	cfg          *config.Config
}

func NewMiddleware(ts *service.TokenService, us *service.UserService, cfg *config.Config) *Middleware {
	return &Middleware{tokenService: ts, userService: us, cfg: cfg}
}

// AuthMiddleware проверяет JWT access token и помещает userID в контекст.
func (m *Middleware) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
		}

		tokenString := headerParts[1]
		claims, err := m.tokenService.ValidateToken(tokenString, "access", m.cfg.JWTAccessSecret)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Set("userID", claims.UserID)
		return next(c)
	}
}

// AdminMiddleware проверяет, является ли пользователь администратором.
// ВАЖНО: Этот middleware должен выполняться ПОСЛЕ AuthMiddleware.
func (m *Middleware) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("userID").(uint)
		if !ok {
			// Этого не должно произойти, если AuthMiddleware отработал правильно
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "user id not found in context"})
		}

		user, err := m.userService.GetByID(userID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}

		if user.Role != models.AdminRole {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied: user is not an admin"})
		}

		return next(c)
	}
}