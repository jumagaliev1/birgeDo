package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserTasks struct {
	UserID int
	User   string
	Task   *[]Task
}
type UserTask struct {
	UserID int
	User   string
	Task   string
	Done   bool
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
			INSERT INTO users (name, email, password_hash, activated)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, version`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
func (m UserModel) GetAll() ([]User, error) {
	query := `
 		SELECT id, name, email
		FROM users`

	var users []User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "rooms_users_pkey"`:
			return nil, ErrDuplicateKey
		default:
			return nil, err
		}
	}

	for rows.Next() {
		var user User
		rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
		)
		users = append(users, user)
	}
	return users, nil
}
func (m UserModel) Get(id int) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE id = $1`
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetRoomsByUser(id int) ([]Room, error) {
	query := `
		SELECT r.id,r.title FROM rooms r 
		INNER JOIN rooms_users ru ON r.id = ru.room_id AND ru.user_id = $1;
	`
	var rooms []Room
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)
	for rows.Next() {
		var room Room
		err = rows.Scan(&room.ID, &room.Title)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound
			default:
				return nil, err
			}
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (m UserModel) GetTasksByUser(id int) ([]Task, error) {
	query := `
		SELECT t.id, t.title, t.room_id, ut.done from tasks t
		INNER JOIN users_tasks ut ON t.id = ut.task_id AND ut.user_id = $1`

	var tasks []Task

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	fmt.Println("no dta")
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
		err = rows.Scan(&task.ID, &task.Title, &task.RoomID, &task.Done)
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

func (m UserModel) InsertRoomUser(userID, roomID int) error {
	query := `
		INSERT INTO rooms_users (user_id, room_id) 
		VALUES ($1, $2)`

	args := []interface{}{userID, roomID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "rooms_users_pkey"`:
			return ErrDuplicateKey
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) RemoveRoomUser(userID, roomID int) error {
	query := `DELETE FROM users_tasks 
				WHERE user_id = $1 AND task_id IN (SELECT id FROM tasks WHERE room_id = $2)`

	args := []interface{}{userID, roomID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	query = `DELETE FROM rooms_users 
				WHERE user_id = $1 AND room_id = $2`
	_, err = m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) RemoveUserTask(taskID, roomID int) error {
	query := `DELETE FROM tasks 
				WHERE id = $1 AND room_id = $2`

	args := []interface{}{taskID, roomID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (m UserModel) GetUsersByRoom(roomID int) ([]int, error) {
	query := `
		SELECT user_id FROM rooms_users
		WHERE room_id = $1`

	var users []int
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		users = append(users, id)
	}

	return users, nil
}

func (m UserModel) InsertUserTask(userID, taskID int) error {
	query := `
		INSERT INTO users_tasks (user_id, task_id, done)
		VALUES ($1, $2, $3)`

	args := []interface{}{userID, taskID, false}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (m UserModel) GetUserTask(roomID int64) ([]UserTask, error) {
	query := `
		SELECT u.id, u.name, t.title, ut.done FROM users_tasks ut 
		    JOIN users u ON u.id = ut.user_id 
		    JOIN tasks t ON t.id = ut.task_id
		    AND t.room_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var usersTasks []UserTask
	rows, err := m.DB.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var userTask UserTask
		err = rows.Scan(&userTask.UserID, &userTask.User, &userTask.Task, &userTask.Done)
		if err != nil {
			return nil, err
		}
		usersTasks = append(usersTasks, userTask)

	}
	return usersTasks, nil
}

func (m UserModel) GetUserTaskByBothID(userID int, taskID int64) (*UserTask, error) {
	query := `
		SELECT user_id, task_id, done FROM users_tasks 
		WHERE user_id = $1 and task_id = $2`
	args := []interface{}{userID, taskID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var userTask UserTask
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&userTask.User, &userTask.Task, &userTask.Done)
	if err != nil {
		return nil, err
	}
	return &userTask, err
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
			SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
			FROM users
			INNER JOIN tokens
			ON users.id = tokens.user_id
			WHERE tokens.hash = $1
			AND tokens.scope = $2
			AND tokens.expiry > $3`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
