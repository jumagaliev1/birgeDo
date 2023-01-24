package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Task struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	RoomID int64  `json:"room_id"`
	Done   bool   `json:"done"`
}

type TaskModel struct {
	DB *sql.DB
}

func (m TaskModel) Insert(task *Task) (int, error) {
	query := `
			INSERT INTO tasks (title, room_id)
			VALUES ($1, $2)
			RETURNING id`
	args := []interface{}{task.Title, task.RoomID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&task.ID)
	if err != nil {
		return 0, err
	}
	return int(task.ID), nil
}
func (m TaskModel) GetByID(id int64) (*Task, error) {
	query := `
			SELECT id, title, room_id 
			FROM tasks
			WHERE id = $1`
	var task Task

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.RoomID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err

		}
	}
	return &task, nil
}

func (m TaskModel) Update(task *Task) error {
	query := `
		UPDATE tasks
		SET title = $1, room_id = $2
		WHERE id = $3`

	args := []interface{}{
		task.Title,
		task.RoomID,
		task.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m TaskModel) GetByRoomID(id int64) ([]Task, error) {
	query := `
			SELECT id, title, room_id 
			FROM tasks
			WHERE room_id = $1`
	var tasks []Task

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err

		}
	}
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Title, &task.RoomID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound
			default:
				return nil, err

			}
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m TaskModel) UpdateUserTaskByBothIDFalse(userID int, taskID int) error {
	query := `UPDATE users_tasks
			SET done = false
			WHERE user_id = $1 and task_id = $2`
	args := []interface{}{
		userID,
		taskID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil

}

func (m TaskModel) UpdateUserTaskByBothIDTrue(userID int, taskID int) error {
	query := `UPDATE users_tasks
			SET done = true
			WHERE user_id = $1 and task_id = $2`
	args := []interface{}{
		userID,
		taskID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil

}
