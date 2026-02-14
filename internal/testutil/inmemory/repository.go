package inmemory

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

// Repository is a partial in-memory implementation of repository.Querier.
// It is designed for handler/service tests and can be extended method-by-method.
type Repository struct {
	repository.Querier

	mu sync.RWMutex

	users map[uuid.UUID]repository.User

	usersByEmail  map[string]repository.GetUserCredentialsByEmailForAuthRow
	refreshTokens map[string]repository.RefreshToken

	students          map[uuid.UUID]repository.Student
	classrooms        map[uuid.UUID]repository.Classroom
	studentClassrooms map[string]repository.StudentClassroom

	rules       map[uuid.UUID]repository.Rule
	bonuses     map[uuid.UUID]repository.Bonus
	penalties   map[uuid.UUID]repository.Penalty
	punishments map[uuid.UUID]repository.Punishment

	bonusTypes      map[uuid.UUID]repository.BonusType
	penaltyTypes    map[uuid.UUID]repository.PenaltyType
	punishmentTypes map[uuid.UUID]repository.PunishmentType

	errors map[string]error
}

func NewRepository() *Repository {
	return &Repository{
		users: make(map[uuid.UUID]repository.User),

		usersByEmail:  make(map[string]repository.GetUserCredentialsByEmailForAuthRow),
		refreshTokens: make(map[string]repository.RefreshToken),

		students:          make(map[uuid.UUID]repository.Student),
		classrooms:        make(map[uuid.UUID]repository.Classroom),
		studentClassrooms: make(map[string]repository.StudentClassroom),

		rules:       make(map[uuid.UUID]repository.Rule),
		bonuses:     make(map[uuid.UUID]repository.Bonus),
		penalties:   make(map[uuid.UUID]repository.Penalty),
		punishments: make(map[uuid.UUID]repository.Punishment),

		bonusTypes:      make(map[uuid.UUID]repository.BonusType),
		penaltyTypes:    make(map[uuid.UUID]repository.PenaltyType),
		punishmentTypes: make(map[uuid.UUID]repository.PunishmentType),

		errors: make(map[string]error),
	}
}

func (r *Repository) SetError(op string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors[op] = err
}

func (r *Repository) ClearError(op string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.errors, op)
}

func (r *Repository) errFor(op string) error {
	return r.errors[op]
}

func (r *Repository) setAuthUser(id uuid.UUID, email, passwordHash string) {
	now := time.Now()
	email = strings.ToLower(email)

	r.users[id] = repository.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	r.usersByEmail[strings.ToLower(email)] = repository.GetUserCredentialsByEmailForAuthRow{
		ID:           id,
		Email:        strings.ToLower(email),
		PasswordHash: passwordHash,
	}
}

func studentClassroomKey(studentID, classroomID uuid.UUID) string {
	return studentID.String() + ":" + classroomID.String()
}
