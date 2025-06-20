package service

import (
	"errors"
	"gorm.io/gorm"
	"server/internal/config"
	"server/internal/models"
	"server/internal/pkg/hash"
	"server/internal/repository/postgres"
)

type AuthService struct {
	userRepo     *postgr.UserRepository
	tokenService *TokenService
	cfg          *config.Config
}

func NewAuthService(userRepo *postgr.UserRepository, tokenService *TokenService, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, tokenService: tokenService, cfg: cfg}
}

func (s *AuthService) Register(dto models.CreateUserDTO) (*models.User, error) {
	hashedPassword, err := hash.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	role := dto.Role
	if role != models.AdminRole && role != models.UserRole {
		role = models.UserRole
	}

	user := &models.User{
		Name:     dto.Name,
		Email:    dto.Email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(dto models.LoginDTO) (string, string, error) {
	user, err := s.userRepo.GetByEmail(dto.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	if !hash.CheckPasswordHash(dto.Password, user.Password) {
		return "", "", errors.New("invalid credentials")
	}

	return s.tokenService.GenerateTokens(user)
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := s.tokenService.ValidateToken(refreshToken, "refresh", s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", "", err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", "", err
	}

	return s.tokenService.GenerateTokens(user)
}