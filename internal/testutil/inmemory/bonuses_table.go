package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateBonus           = "CreateBonus"
	OpGetBonusByUser        = "GetBonusByUser"
	OpCountBonusesByUser    = "CountBonusesByUser"
	OpListBonusesByUser     = "ListBonusesByUser"
	OpCountBonusesByStudent = "CountBonusesByStudent"
	OpListBonusesByStudent  = "ListBonusesByStudent"
	OpUseBonus              = "UseBonus"
	OpDeleteBonusByUser     = "DeleteBonusByUser"
)

func (r *Repository) bonusTypeNameForBonusLocked(bonusTypeID uuid.UUID) string {
	if bonusType, ok := r.bonusTypes[bonusTypeID]; ok {
		return bonusType.Name
	}

	return ""
}

func (r *Repository) buildCreateBonusRowLocked(bonus repository.Bonus) repository.CreateBonusRow {
	studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)

	return repository.CreateBonusRow{
		ID:               bonus.ID,
		UserID:           bonus.UserID,
		StudentID:        bonus.StudentID,
		BonusTypeID:      bonus.BonusTypeID,
		Points:           bonus.Points,
		CreatedAt:        bonus.CreatedAt,
		UsedAt:           bonus.UsedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
	}
}

func (r *Repository) buildGetBonusRowLocked(bonus repository.Bonus) repository.GetBonusByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)

	return repository.GetBonusByUserRow{
		ID:               bonus.ID,
		UserID:           bonus.UserID,
		StudentID:        bonus.StudentID,
		BonusTypeID:      bonus.BonusTypeID,
		Points:           bonus.Points,
		CreatedAt:        bonus.CreatedAt,
		UsedAt:           bonus.UsedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
	}
}

func (r *Repository) buildListBonusByUserRowLocked(bonus repository.Bonus) repository.ListBonusesByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)

	return repository.ListBonusesByUserRow{
		ID:               bonus.ID,
		UserID:           bonus.UserID,
		StudentID:        bonus.StudentID,
		BonusTypeID:      bonus.BonusTypeID,
		Points:           bonus.Points,
		CreatedAt:        bonus.CreatedAt,
		UsedAt:           bonus.UsedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
	}
}

func (r *Repository) buildListBonusByStudentRowLocked(bonus repository.Bonus) repository.ListBonusesByStudentRow {
	studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)

	return repository.ListBonusesByStudentRow{
		ID:               bonus.ID,
		UserID:           bonus.UserID,
		StudentID:        bonus.StudentID,
		BonusTypeID:      bonus.BonusTypeID,
		Points:           bonus.Points,
		CreatedAt:        bonus.CreatedAt,
		UsedAt:           bonus.UsedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
	}
}

func (r *Repository) buildUseBonusRowLocked(bonus repository.Bonus) repository.UseBonusRow {
	studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)

	return repository.UseBonusRow{
		ID:               bonus.ID,
		UserID:           bonus.UserID,
		StudentID:        bonus.StudentID,
		BonusTypeID:      bonus.BonusTypeID,
		Points:           bonus.Points,
		CreatedAt:        bonus.CreatedAt,
		UsedAt:           bonus.UsedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
	}
}

func (r *Repository) SeedBonus(bonus repository.Bonus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if bonus.ID == uuid.Nil {
		bonus.ID = uuid.New()
	}
	if bonus.CreatedAt.IsZero() {
		bonus.CreatedAt = now
	}

	r.bonuses[bonus.ID] = bonus
}

func (r *Repository) CreateBonus(_ context.Context, arg repository.CreateBonusParams) (repository.CreateBonusRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateBonus); err != nil {
		return repository.CreateBonusRow{}, err
	}

	bonus := repository.Bonus{
		ID:          uuid.New(),
		UserID:      arg.UserID,
		StudentID:   arg.StudentID,
		BonusTypeID: arg.BonusTypeID,
		Points:      arg.Points,
		CreatedAt:   time.Now(),
	}
	r.bonuses[bonus.ID] = bonus

	return r.buildCreateBonusRowLocked(bonus), nil
}

func (r *Repository) GetBonusByUser(_ context.Context, arg repository.GetBonusByUserParams) (repository.GetBonusByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetBonusByUser); err != nil {
		return repository.GetBonusByUserRow{}, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID {
		return repository.GetBonusByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetBonusRowLocked(bonus), nil
}

func (r *Repository) CountBonusesByUser(_ context.Context, arg repository.CountBonusesByUserParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountBonusesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListBonusesByUser(_ context.Context, arg repository.ListBonusesByUserParams) ([]repository.ListBonusesByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListBonusesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			items = append(items, bonus)
		}
	}

	sortBonusesByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListBonusesByUserRow, 0, len(paginated))
	for _, bonus := range paginated {
		rows = append(rows, r.buildListBonusByUserRowLocked(bonus))
	}

	return rows, nil
}

func (r *Repository) CountBonusesByStudent(_ context.Context, arg repository.CountBonusesByStudentParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountBonusesByStudent); err != nil {
		return 0, err
	}

	var count int64
	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListBonusesByStudent(_ context.Context, arg repository.ListBonusesByStudentParams) ([]repository.ListBonusesByStudentRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListBonusesByStudent); err != nil {
		return nil, err
	}

	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			items = append(items, bonus)
		}
	}

	sortBonusesByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListBonusesByStudentRow, 0, len(paginated))
	for _, bonus := range paginated {
		rows = append(rows, r.buildListBonusByStudentRowLocked(bonus))
	}

	return rows, nil
}

func (r *Repository) UseBonus(_ context.Context, arg repository.UseBonusParams) (repository.UseBonusRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUseBonus); err != nil {
		return repository.UseBonusRow{}, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID || bonus.UsedAt.Valid {
		return repository.UseBonusRow{}, pgx.ErrNoRows
	}

	bonus.UsedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	r.bonuses[arg.ID] = bonus

	return r.buildUseBonusRowLocked(bonus), nil
}

func (r *Repository) DeleteBonusByUser(_ context.Context, arg repository.DeleteBonusByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteBonusByUser); err != nil {
		return 0, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.bonuses, arg.ID)
	return 1, nil
}
