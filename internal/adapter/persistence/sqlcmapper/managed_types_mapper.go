package sqlcmapper

import (
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func BonusTypeFromRepository(bt *repository.BonusType) *dto.ReturnBonusTypeDto {
	if bt == nil {
		return nil
	}

	return &dto.ReturnBonusTypeDto{
		ID:        bt.ID,
		Name:      bt.Name,
		CreatedAt: bt.CreatedAt,
		UpdatedAt: bt.UpdatedAt,
	}
}

func PenaltyTypeFromRepository(pt *repository.PenaltyType) *dto.ReturnPenaltyTypeDto {
	if pt == nil {
		return nil
	}

	return &dto.ReturnPenaltyTypeDto{
		ID:        pt.ID,
		Name:      pt.Name,
		CreatedAt: pt.CreatedAt,
		UpdatedAt: pt.UpdatedAt,
	}
}

func PunishmentTypeFromRepository(pt *repository.PunishmentType) *dto.ReturnPunishmentTypeDto {
	if pt == nil {
		return nil
	}

	return &dto.ReturnPunishmentTypeDto{
		ID:        pt.ID,
		Name:      pt.Name,
		CreatedAt: pt.CreatedAt,
		UpdatedAt: pt.UpdatedAt,
	}
}

func UserFromRepository(u *repository.CreateUserRow) *dto.ReturnUserDto {
	if u == nil {
		return nil
	}

	return &dto.ReturnUserDto{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func UserFromGetByIDRow(u *repository.GetUserByIDRow) *dto.ReturnUserDto {
	if u == nil {
		return nil
	}

	return &dto.ReturnUserDto{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
