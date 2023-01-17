package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	//ErrInvalidCredentials = errors.New("models: invalid credentials")
)

type Models struct {
	Users UserModel
	Task  TaskModel
	Room  RoomModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{
			DB: db,
		},
		Task: TaskModel{
			DB: db,
		},
		Room: RoomModel{DB: db},
	}

}
