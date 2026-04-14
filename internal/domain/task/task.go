package task

import (
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type ScheduleType string

const (
	ScheduleNone    ScheduleType = ""
	ScheduleDaily   ScheduleType = "daily"
	ScheduleMonthly ScheduleType = "monthly"
	ScheduleDates   ScheduleType = "dates"
	ScheduleParity  ScheduleType = "parity"
)

type Schedule struct {
	Type     ScheduleType `json:"type"`
	Interval int          `json:"interval,omitempty"` // для daily
	Day      int          `json:"day,omitempty"`      // для monthly
	Dates    []string     `json:"dates,omitempty"`    // для dates (формат "2006-01-02")
	Parity   string       `json:"parity,omitempty"`   // "even" или "odd"
}

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Schedule    *Schedule `json:"schedule"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (s *Schedule) NextExecution(base time.Time, now time.Time) *time.Time {
	if s == nil || s.Type == ScheduleNone {
		return nil
	}

	switch s.Type {
	case ScheduleDaily:
		next := base.AddDate(0, 0, s.Interval)
		for next.Before(now) {
			next = next.AddDate(0, 0, s.Interval)
		}
		return &next

	case ScheduleMonthly:
		year, month, _ := base.Date()
		loc := base.Location()
		candidate := time.Date(year, month, s.Day, 0, 0, 0, 0, loc)
		if candidate.Day() != s.Day {
			candidate = time.Date(year, month+1, 0, 0, 0, 0, 0, loc)
		}
		for candidate.Before(now) {
			candidate = candidate.AddDate(0, 1, 0)
			if candidate.Day() != s.Day {
				candidate = time.Date(candidate.Year(), candidate.Month()+1, 0, 0, 0, 0, 0, loc)
			}
		}
		return &candidate

	case ScheduleDates:
		nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for _, dateStr := range s.Dates {
			t, _ := time.Parse("2006-01-02", dateStr)
			if !t.Before(nowDay) {
				return &t
			}
		}
		return nil

	case ScheduleParity:
		candidate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for {
			day := candidate.Day()
			isEven := day%2 == 0
			if (s.Parity == "even" && isEven) || (s.Parity == "odd" && !isEven) {
				return &candidate
			}
			candidate = candidate.AddDate(0, 0, 1)
		}

	default:
		return nil
	}
}
