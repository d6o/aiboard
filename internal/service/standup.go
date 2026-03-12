package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/d6o/aiboard/internal/model"
)

type standupStore interface {
	GetConfig() (model.StandupConfig, error)
	SaveConfig(intervalHours int, enabled bool) (model.StandupConfig, error)
	FindAll(limit int) ([]model.Standup, error)
	FindByID(id string) (model.Standup, error)
	FindLatest() (model.Standup, error)
	Create(number int, startTime, endTime time.Time) (model.Standup, error)
	FindEntriesByStandupID(standupID string) ([]model.StandupEntry, error)
	CreateEntry(standupID, userID, content string) (model.StandupEntry, error)
	NextNumber() (int, error)
}

type standupUserFinder interface {
	FindAll() ([]model.User, error)
}

type standupNotifier interface {
	Create(userID, message, cardID string) (model.Notification, error)
}

type StandupService struct {
	standups standupStore
	users    standupUserFinder
	notifs   standupNotifier
}

func NewStandupService(standups standupStore, users standupUserFinder, notifs standupNotifier) *StandupService {
	return &StandupService{
		standups: standups,
		users:    users,
		notifs:   notifs,
	}
}

func (s *StandupService) GetConfig() (model.StandupConfig, error) {
	return s.standups.GetConfig()
}

func (s *StandupService) UpdateConfig(intervalHours int, enabled bool) (model.StandupConfig, error) {
	if intervalHours < 1 {
		return model.StandupConfig{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "interval_hours", Message: "interval_hours must be at least 1"}},
		}
	}
	return s.standups.SaveConfig(intervalHours, enabled)
}

func (s *StandupService) List(limit int) ([]model.Standup, error) {
	return s.standups.FindAll(limit)
}

func (s *StandupService) Get(id string) (model.Standup, error) {
	st, err := s.standups.FindByID(id)
	if err != nil {
		return st, err
	}

	entries, err := s.standups.FindEntriesByStandupID(id)
	if err != nil {
		return st, err
	}
	st.Entries = entries
	return st, nil
}

func (s *StandupService) PostEntry(standupID, userID, content string) (model.StandupEntry, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return model.StandupEntry{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "content", Message: "content is required"}},
		}
	}

	if _, err := s.standups.FindByID(standupID); err != nil {
		return model.StandupEntry{}, err
	}

	return s.standups.CreateEntry(standupID, userID, content)
}

func (s *StandupService) RunTicker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndCreateStandup()
		}
	}
}

func (s *StandupService) checkAndCreateStandup() {
	cfg, err := s.standups.GetConfig()
	if err != nil || !cfg.Enabled {
		return
	}

	now := time.Now()
	latest, err := s.standups.FindLatest()
	if err != nil {
		if err != model.ErrNotFound {
			return
		}
		// No standups yet — create the first one
		s.createStandup(cfg, now)
		return
	}

	// Check if enough time has passed since the last standup's end time
	if now.Before(latest.EndTime) {
		return
	}

	s.createStandup(cfg, now)
}

func (s *StandupService) createStandup(cfg model.StandupConfig, now time.Time) {
	num, err := s.standups.NextNumber()
	if err != nil {
		return
	}

	interval := time.Duration(cfg.IntervalHours) * time.Hour
	startTime := now
	endTime := now.Add(interval)

	st, err := s.standups.Create(num, startTime, endTime)
	if err != nil {
		log.Println("failed to create standup:", err)
		return
	}

	s.notifyAllUsers(st)
}

func (s *StandupService) notifyAllUsers(st model.Standup) {
	allUsers, err := s.users.FindAll()
	if err != nil {
		return
	}

	startStr := st.StartTime.Format("Jan 2 15:04")
	endStr := st.EndTime.Format("Jan 2 15:04")
	msg := "StandUp #" + formatNumber(st.Number) +
		" is open! Post what you did between " + startStr + " and " + endStr

	for _, u := range allUsers {
		s.notifs.Create(u.ID, msg, "")
	}
}

func formatNumber(n int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if s == "" {
		return "0"
	}
	return s
}
