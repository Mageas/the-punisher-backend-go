package handler

import "github.com/mageas/the-punisher-backend/internal/service"

type ClassroomHandler struct {
	service service.ClassroomService
}

func NewClassroomHandler(service service.ClassroomService) *ClassroomHandler {
	return &ClassroomHandler{service: service}
}
