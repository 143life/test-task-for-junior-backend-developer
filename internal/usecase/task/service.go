package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Schedule:    normalized.Schedule,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Schedule:    normalized.Schedule,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Schedule != nil {
		if err := validateSchedule(input.Schedule); err != nil {
			return CreateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Schedule != nil {
		if err := validateSchedule(input.Schedule); err != nil {
			return UpdateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	return input, nil
}

func validateSchedule(s *taskdomain.Schedule) error {
	switch s.Type {
	case taskdomain.ScheduleNone:
		return errors.New("schedule type cannot be empty")
	case taskdomain.ScheduleDaily:
		if s.Interval <= 0 {
			return errors.New("daily schedule requires interval > 0")
		}
	case taskdomain.ScheduleMonthly:
		if s.Day < 1 || s.Day > 31 {
			return errors.New("monthly schedule requires day between 1 and 31")
		}
	case taskdomain.ScheduleDates:
		if len(s.Dates) == 0 {
			return errors.New("dates schedule requires at least one date")
		}
		for _, d := range s.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return fmt.Errorf("invalid date format in dates schedule: %s", d)
			}
		}
	case taskdomain.ScheduleParity:
		if s.Parity != "even" && s.Parity != "odd" {
			return errors.New("parity schedule requires 'even' or 'odd'")
		}
	default:
		return fmt.Errorf("unknown schedule type: %s", s.Type)
	}
	return nil
}
