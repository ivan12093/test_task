package profile

import "server/internal/domain"

type profileDTO struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (dto *profileDTO) ToDomain() *domain.Profile {
	return &domain.Profile{
		FullName: dto.FullName,
		Phone:    dto.Phone,
		Email:    dto.Email,
	}
}

func (dto *profileDTO) FromDomain(profile *domain.Profile) {
	dto.FullName = profile.FullName
	dto.Phone = profile.Phone
	dto.Email = profile.Email
}
