package db

import (
	"github.com/jackc/pgx"
)

type UserModel struct {
	Id                string
	Username          *string
	Email             string
	EncryptedPassword []byte
}

var PgPool *pgx.ConnPool

func init() {
	PgPool, _ = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "db",
			User:     "postgres",
			Password: "postgres",
			Database: "mg4",
		},
	})
}

// CreateUser takes a UserModel and writes it to the database.
// If this write was successful, it returns a Usermodel as seen by the database and a nil error.
// Otherwise, it returns a nil model and an errror
func CreateUser(u UserModel) (*UserModel, error) {
	const qsIns = "INSERT INTO users(username, email, encrypted_password) VALUES($1, $2, $3)"
	const qsSel = "SELECT id FROM users WHERE username=$1 AND email=$2"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)


	// Attempt to insert the new user
	if _, err = conn.Exec(qsIns, u.Username, u.Email, u.EncryptedPassword); err != nil {
		return nil, err
	}

	// Attempt to find the new user's id by username and email
	row := conn.QueryRow(qsSel, u.Username, u.Email)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}
	u.Id = id

	return &u, nil
}

// DeleteUser takes a user id and removes the row containing that user from the database/
// If this delete was sucessful, it returns nil.
// Otherwise it returns an error
func DeleteUser(id string) error {
	const qs = "DELETE FROM users WHERE id=$1"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return err
	}
	defer PgPool.Release(conn)

	// Attempt to delete the user by id
	if _, err = conn.Exec(qs, id); err != nil {
		return err
	}
	return nil
}

func ListUsers(ids []string) (*[]UserModel, error) {
	const qsIn = "SELECT * FROM users WHERE id IN $1"
	const qsAll = "SELECT * FROM users"
	return nil, nil
}

func GetUserById(id string) (*UserModel, error) {
	const qs = "SELECT username, email, encrypted_password FROM users WHERE id=$1"
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var username *string
	var email string
	var encrypted_password []byte
	row := conn.QueryRow(qs, id)
	err = row.Scan(&username, &email, &encrypted_password)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		Id:                id,
		Username:          username,
		Email:             email,
		EncryptedPassword: encrypted_password,
	}, nil
}

func GetUserByEmail(email string) (*UserModel, error) {
	const qs = "SELECT id, username, encrypted_password FROM users WHERE email=$1"
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var id string
	var username *string
	var encrypted_password []byte
	row := conn.QueryRow(qs, email)
	err = row.Scan(&id, &username, &encrypted_password)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		Id:                id,
		Username:          username,
		Email:             email,
		EncryptedPassword: encrypted_password,
	}, nil
}

func GetUserByUsername(username string) (*UserModel, error) {
	const qs = "SELECT id, email, encrypted_password FROM users WHERE username=$1"
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var id string
	var email string
	var encrypted_password []byte
	row := conn.QueryRow(qs, username)
	err = row.Scan(&id, &email, &encrypted_password)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		Id:                id,
		Username:          &username,
		Email:             email,
		EncryptedPassword: encrypted_password,
	}, nil
}

func UpdateUser(u UserModel) error {
	const qs = "UPDATE users SET username=$2, email=$3, encrypted_password=$4 WHERE id=$1"
	return nil
}

// CreateUserByFacebook takes a facebookId and an email and creates a new user with that email, then links the
// facebook_users table to that new user
func CreateUserByFacebook(facebookId string, email string) (*UserModel, error) {
	const qsInsUser = "INSERT INTO users(email) VALUES($1)"
	const qsSel = "SELECT id, username FROM users where email=$1"
	const qsInsFBUser = "INSERT INTO facebook_users(facebook_user_id, user_id) VALUES ($1, $2)"

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	// Attempt to insert the new user
	if _, err = conn.Exec(qsInsUser, email); err != nil {
		return nil, err
	}

	// Attempt to find the new user's id by username and email
	row := conn.QueryRow(qsSel, email)
	var id string
	var username *string
	if err = row.Scan(&id, &username); err != nil {
		return nil, err
	}

	// Attempt to write the facebook id link to facebook_users
	if _, err = conn.Exec(qsInsFBUser, facebookId, id); err != nil {
		return nil, err
	}

	return &UserModel{
		Id:       id,
		Username: username,
		Email:    email,
	}, nil
}

// GetUserByFacebook takes a facebook user id (provided by facebook per app) and uses it to look for linked users
// If a link exists, it retrieves the user by the id linked.
func GetUserByFacebook(facebookId string) (*UserModel, error) {
	const qs = "SELECT id, username, email FROM users WHERE id=(SELECT user_id FROM facebook_users WHERE facebook_user_id=$1)"
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var id string
	var username *string
	var email string
	row := conn.QueryRow(qs, facebookId)
	err = row.Scan(&id, &username, &email)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		Id:       id,
		Username: username,
		Email:    email,
	}, nil
}
