package scheduler

import (
	"context"
	"log"
	"strconv"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository — минимальный интерфейс репозитория, нужный планировщику.
type Repository interface {
	ListTasksWithSchedule(ctx context.Context) ([]taskdomain.Task, error)
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*taskdomain.Task, error)
}

type Scheduler struct {
	repo   Repository
	pool   *pgxpool.Pool
	done   chan struct{}
	logger *log.Logger
}

func New(repo Repository, pool *pgxpool.Pool, logger *log.Logger) *Scheduler {
	return &Scheduler{
		repo:   repo,
		pool:   pool,
		done:   make(chan struct{}),
		logger: logger,
	}
}

// Start запускает планировщик: выполняет начальное полное сканирование и подписывается на уведомления.
func (s *Scheduler) Start(ctx context.Context) {
	// Начальное сканирование при старте (чтобы не пропустить задачи, пропущенные во время простоя)
	go func() {
		time.Sleep(2 * time.Second) // даём БД прогреться
		s.logger.Println("scheduler: initial full scan started")
		s.scanAll(ctx)
		s.logger.Println("scheduler: initial full scan completed")
	}()

	// Запускаем слушатель уведомлений
	go s.listen(ctx)
}

// Stop останавливает планировщик.
func (s *Scheduler) Stop() {
	close(s.done)
}

// listen подписывается на канал task_schedule_changed и обрабатывает уведомления.
func (s *Scheduler) listen(ctx context.Context) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.logger.Printf("scheduler: acquire connection for listen: %v", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN task_schedule_changed")
	if err != nil {
		s.logger.Printf("scheduler: LISTEN failed: %v", err)
		return
	}

	s.logger.Println("scheduler: listening on task_schedule_changed")

	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		default:
		}

		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Printf("scheduler: WaitForNotification error: %v", err)
			time.Sleep(5 * time.Second) // пауза перед переподключением
			continue
		}

		taskID, err := strconv.ParseInt(notification.Payload, 10, 64)
		if err != nil {
			s.logger.Printf("scheduler: invalid payload %q: %v", notification.Payload, err)
			continue
		}

		s.logger.Printf("scheduler: processing task %d from notification", taskID)
		s.processTask(ctx, taskID)
	}
}

// scanAll проходит по всем задачам с schedule и создаёт дочерние при необходимости.
func (s *Scheduler) scanAll(ctx context.Context) {
	tasks, err := s.repo.ListTasksWithSchedule(ctx)
	if err != nil {
		s.logger.Printf("scheduler: failed to list tasks: %v", err)
		return
	}

	now := time.Now().UTC()
	for i := range tasks {
		s.createChildIfNeeded(ctx, &tasks[i], now)
	}
}

// processTask обрабатывает одну задачу по ID.
func (s *Scheduler) processTask(ctx context.Context, id int64) {
	// Используем FOR UPDATE SKIP LOCKED, чтобы только один экземпляр планировщика взял задачу
	t, err := s.repo.GetByIDForUpdate(ctx, id)
	if err != nil {
		if err == taskdomain.ErrNotFound {
			// Задача удалена — ничего не делаем
			return
		}
		s.logger.Printf("scheduler: GetByIDForUpdate %d: %v", id, err)
		return
	}

	now := time.Now().UTC()
	s.createChildIfNeeded(ctx, t, now)
}

// createChildIfNeeded проверяет, нужно ли создать дочернюю задачу, и создаёт её.
func (s *Scheduler) createChildIfNeeded(ctx context.Context, parent *taskdomain.Task, now time.Time) {
	if parent.Schedule == nil {
		return
	}

	base := parent.CreatedAt
	next := parent.Schedule.NextExecution(base, now)
	if next != nil && !next.After(now) {
		child := &taskdomain.Task{
			Title:       parent.Title,
			Description: parent.Description,
			Status:      taskdomain.StatusNew,
			Schedule:    nil, // дочерняя задача одноразовая
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		_, err := s.repo.Create(ctx, child)
		if err != nil {
			s.logger.Printf("scheduler: failed to create child for %d: %v", parent.ID, err)
		} else {
			s.logger.Printf("scheduler: created child for %d", parent.ID)
		}
	}
}
