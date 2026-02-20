package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type ClassroomService interface {
	CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error)
	GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error)
	ListClassrooms(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error)
	UpdateClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, req dto.UpdateClassroomDto) (*dto.ReturnClassroomDto, error)
	DeleteClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) error

	AddStudentToClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error
	RemoveStudentFromClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error
	ListStudentsByClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error)
	ListClassroomsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error)
}

type classroomService struct {
	repo repository.Querier
}

func NewClassroomService(repo repository.Querier) ClassroomService {
	return &classroomService{repo: repo}
}
