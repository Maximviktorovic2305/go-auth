package service

import (
	"server/internal/models"
	"server/internal/pkg/hash"
	"server/internal/repository/postgres"
)

type UserService struct {
	repo *postgr.UserRepository
}

func NewUserService(repo *postgr.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(dto models.CreateUserDTO) (*models.User, error) {
	hashedPassword, err := hash.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	// Default role to "user" if not provided or invalid
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

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetAll() ([]models.User, error) {
	return s.repo.GetAll()
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) Update(id uint, dto models.UpdateUserDTO) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Name != "" {
		user.Name = dto.Name
	}
	if dto.Email != "" {
		user.Email = dto.Email
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(id uint) error {
	return s.repo.Delete(id)
}