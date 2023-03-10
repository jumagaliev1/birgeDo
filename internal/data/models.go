package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateKey   = errors.New("duplicate key")
	//ErrInvalidCredentials = errors.New("models: invalid credentials")
)

type Models struct {
	Users  UserModel
	Task   TaskModel
	Room   RoomModel
	Tokens TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{
			DB: db,
		},
		Task: TaskModel{
			DB: db,
		},
		Room:   RoomModel{DB: db},
		Tokens: TokenModel{DB: db},
	}

}

type InputAddUser struct {
	RoomID int `json:"roomID"`
	UserID int `json:"userID"`
}

type InputRemoveTask struct {
	TaskID int `json:"taskID"`
	RoomID int `json:"roomID"`
}

type InputRegisterUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type InputAuthUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type InputCreateRoom struct {
	Title string `json:"title"`
}

type InputCreateTask struct {
	Title  string `json:"title"`
	RoomID int64  `json:"roomID"`
}
