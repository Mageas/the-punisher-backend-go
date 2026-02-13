package testutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

// MockQuerier implements repository.Querier with in-memory storage.
type MockQuerier struct {
	Students *MemoryStore[repository.Student]
}

func NewMockQuerier() *MockQuerier {
	return &MockQuerier{
		Students: NewMemoryStore(func(s repository.Student) uuid.UUID {
			return s.UserID
		}),
	}
}

// --- Student methods ---

func (m *MockQuerier) CreateStudent(ctx context.Context, arg repository.CreateStudentParams) (repository.Student, error) {
	now := time.Now().UTC()
	s := repository.Student{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		FirstName: arg.FirstName,
		LastName:  arg.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.Students.Set(s.ID, s)
	return s, nil
}

func (m *MockQuerier) GetStudentByUser(ctx context.Context, arg repository.GetStudentByUserParams) (repository.Student, error) {
	return m.Students.GetByIDAndOwner(arg.ID, arg.UserID)
}

func (m *MockQuerier) CountStudentsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return m.Students.CountByOwner(userID), nil
}

func (m *MockQuerier) ListStudentsByUser(ctx context.Context, arg repository.ListStudentsByUserParams) ([]repository.Student, error) {
	return m.Students.ListByOwner(arg.UserID, int(arg.QueryOffset), int(arg.QueryLimit)), nil
}

func (m *MockQuerier) UpdateStudentByUser(ctx context.Context, arg repository.UpdateStudentByUserParams) (repository.Student, error) {
	s, err := m.Students.GetByIDAndOwner(arg.ID, arg.UserID)
	if err != nil {
		return repository.Student{}, err
	}

	if arg.FirstName.Valid {
		s.FirstName = arg.FirstName.String
	}
	if arg.LastName.Valid {
		s.LastName = arg.LastName.String
	}
	s.UpdatedAt = time.Now().UTC()
	m.Students.Set(s.ID, s)
	return s, nil
}

func (m *MockQuerier) DeleteStudentByUser(ctx context.Context, arg repository.DeleteStudentByUserParams) error {
	m.Students.Delete(arg.ID)
	return nil
}

// --- Unused interface methods (not needed for student tests) ---

func (m *MockQuerier) CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error) {
	return repository.RefreshToken{}, nil
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.CreateUserRow, error) {
	return repository.CreateUserRow{}, nil
}

func (m *MockQuerier) DeleteRefreshToken(ctx context.Context, token string) error {
	return nil
}

func (m *MockQuerier) GetRefreshToken(ctx context.Context, arg repository.GetRefreshTokenParams) (repository.RefreshToken, error) {
	return repository.RefreshToken{}, nil
}

func (m *MockQuerier) GetUserCredentialsByEmailForAuth(ctx context.Context, email string) (repository.GetUserCredentialsByEmailForAuthRow, error) {
	return repository.GetUserCredentialsByEmailForAuthRow{}, nil
}

func (m *MockQuerier) ListRefreshTokensByUserId(ctx context.Context, userID uuid.UUID) ([]repository.RefreshToken, error) {
	return nil, nil
}

func (m *MockQuerier) RevokeRefreshToken(ctx context.Context, token string) (repository.RevokeRefreshTokenRow, error) {
	return repository.RevokeRefreshTokenRow{}, nil
}

func (m *MockQuerier) UserEmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}

// Compile-time check
var _ repository.Querier = (*MockQuerier)(nil)

// SeedStudent inserts a student directly into the mock for test setup.
func (m *MockQuerier) SeedStudent(userID uuid.UUID, firstName, lastName string) repository.Student {
	now := time.Now().UTC()
	s := repository.Student{
		ID:        uuid.New(),
		UserID:    userID,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.Students.Set(s.ID, s)
	return s
}

// Unused import guard for pgtype
var _ = pgtype.Text{}
