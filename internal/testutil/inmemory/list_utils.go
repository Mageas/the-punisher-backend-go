package inmemory

import (
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func sortStudentsByCreatedAtDesc(items []repository.Student) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortClassroomsByCreatedAtDesc(items []repository.Classroom) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortRulesByCreatedAtDesc(items []repository.Rule) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortBonusesByCreatedAtDesc(items []repository.Bonus) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortPenaltiesByCreatedAtDesc(items []repository.Penalty) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortPunishmentsByCreatedAtDesc(items []repository.Punishment) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortBonusTypesByCreatedAtDesc(items []repository.BonusType) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortPenaltyTypesByCreatedAtDesc(items []repository.PenaltyType) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func sortPunishmentTypesByCreatedAtDesc(items []repository.PunishmentType) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func paginate[T any](items []T, offset, limit int32) []T {
	if limit <= 0 {
		return []T{}
	}
	if offset < 0 {
		offset = 0
	}
	if int(offset) >= len(items) {
		return []T{}
	}

	end := int(offset + limit)
	if end > len(items) {
		end = len(items)
	}

	return items[int(offset):end]
}

func matchesOptionalBool(filter pgtype.Bool, value bool) bool {
	if !filter.Valid {
		return true
	}

	return filter.Bool == value
}
