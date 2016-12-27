package db

import (
	"github.com/jackc/pgx"
)

type UserModel struct {
	Id                string
	Username          string
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

	// Begin a transaction and set it up to rollback by default
	tx, err := conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Attempt to insert the new user
	if _, err = tx.Exec(qsIns, u.Username, u.Email, u.EncryptedPassword); err != nil {
		return nil, err
	}

	// Attempt to find the new user's id by username and email
	row := tx.QueryRow(qsSel, u.Username, u.Email)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}
	u.Id = id
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &u, nil
}

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

	var username string
	var email string
	var encrypted_password []byte
	row := conn.QueryRow(qs, id)
	err = row.Scan(&username, &email, &encrypted_password)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		Id: id,
		Username: username,
		Email: email,
		EncryptedPassword: encrypted_password,
	}, nil
}

func GetUserByEmail(email string) (*UserModel, error) {
	const qs = "SELECT * FROM users WHERE email=$1"
	return nil, nil
}

func GetUserByUsername(username string) (*UserModel, error) {
	const qs = "SELECT * FROM users WHERE username=$1"
	return nil, nil
}

func UpdateUser(u UserModel) error {
	const qs = "UPDATE users SET username=$2, email=$3, encrypted_password=$4 WHERE id=$1"
	return nil
}
