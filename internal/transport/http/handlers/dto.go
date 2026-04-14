package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      taskdomain.Status    `json:"status"`
	Schedule    *taskdomain.Schedule `json:"schedule,omitempty"`
}

type taskDTO struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      taskdomain.Status    `json:"status"`
	Schedule    *taskdomain.Schedule `json:"schedule,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Schedule:    task.Schedule,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
