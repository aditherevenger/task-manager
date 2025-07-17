package storage

import "task-manager/internal/model"

type TaskStorage interface {

	//save saves all task to storage
	save(tasks []*model.Task) error

	//Load loads all task from storage
	Load() ([]*model.Task, error)
}
