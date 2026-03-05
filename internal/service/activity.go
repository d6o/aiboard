package service

import "github.com/d6o/aiboard/internal/model"

type activityStore interface {
	Find(filter model.ActivityFilter) ([]model.ActivityEntry, error)
}

type ActivityService struct {
	activity activityStore
}

func NewActivityService(activity activityStore) *ActivityService {
	return &ActivityService{activity: activity}
}

func (s *ActivityService) List(filter model.ActivityFilter) ([]model.ActivityEntry, error) {
	return s.activity.Find(filter)
}
