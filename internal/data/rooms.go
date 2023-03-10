package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	"time"
)

type Room struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func ValidateRoom(v *validator.Validator, room *Room) {
	v.Check(room.Title != "", "title", "must be provided")
	v.Check(len(room.Title) > 100, "title", "must be less than 100 bytes long")
}

type RoomModel struct {
	DB *sql.DB
}

func (m RoomModel) Insert(room *Room) (int, error) {
	query := `
		INSERT INTO rooms (title)
		VALUES ($1)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, room.Title).Scan(&room.ID)
	if err != nil {
		return 0, err
	}
	return int(room.ID), nil
}

func (m RoomModel) GetByID(id int64) (*Room, error) {
	query := `
			SELECT id, title 
			FROM rooms
			WHERE id = $1`
	var room Room
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&room.ID,
		&room.Title)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err

		}
	}
	return &room, nil
}

func (m RoomModel) Update(room *Room) error {
	query := `
		UPDATE tasks
		SET title = $1
		WHERE id = $2`

	args := []interface{}{
		room.Title,
		room.ID,
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
