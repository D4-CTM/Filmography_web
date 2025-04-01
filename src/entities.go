package renderer

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type users struct {
	Id       int    `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
	PfpUrl   string `db:"pfp_url"`
	Password int    `db:"password"`
}

func (user *users) Insert(db *sqlx.DB) error {
    _, err := db.Exec(`CALL sp_insert_user(?, ?, ?, ?)`,
        user.Username,
        user.Email,
        user.PfpUrl,
        user.Password)
    if err != nil {
        return fmt.Errorf("Crash when inserting user!\nerr.Error(): %v\n", err.Error())
    }
    return nil
}

func (user *users) Update(db *sqlx.DB) error {
    _, err := db.Exec(`CALL sp_update_user(?, ?, ?, ?, ?)`,
        user.Id,
        user.Username,
        user.Email,
        user.PfpUrl,
        user.Password)
    if err != nil {
        return fmt.Errorf("Crash when updating user!\nerr.Error(): %v\n", err.Error())
    }
    return nil
}

