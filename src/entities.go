package renderer

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Users struct {
	Id       int            `db:"id"`
	Username string         `db:"username"`
	Email    string         `db:"email"`
	PfpUrl   sql.NullString `db:"pfp_url"`
	Password int            `db:"password"`
}

func (user *Users) Insert(db *sqlx.DB) error {
	err := db.QueryRow(`CALL sp_insert_user($1,$2,$3,$4,$5)`,
		&user.Id,
		user.Username,
		user.Email,
		user.PfpUrl,
		user.Password).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("Crash when inserting user!\nerr.Error(): %v\n", err.Error())
	}
	return nil
}

func (user *Users) Update(db *sqlx.DB) error {
	_, err := db.Exec(`CALL sp_update_user($1,$2,$3,$4,$5)`,
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

func (user *Users) Fetch(db *sqlx.DB) error {
    err := db.Get(user, `SELECT * FROM users WHERE username = $1 AND password = $2`, user.Username, user.Password)
    if err != nil {
        return fmt.Errorf("Crash while fetching the user! \nPlease check the username of password typed! \nerr.Error(): %v\n", err.Error())
    }
    return nil
}

type Movies struct {
    Id          int            `db:"id"`
    Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	PosterUrl   sql.NullString `db:"poster_url"`
	Stars       int16          `db:"stars"`
	AddedBy     int            `db:"added_by"`
}

func (movie *Movies) Insert(db *sqlx.DB) error {
	_, err := db.Exec(`CALL sp_insert_movie($1,$2,$3,$4,$5)`,
		movie.Name,
		movie.Description,
		movie.PosterUrl,
		movie.Stars,
		movie.AddedBy)

	if err != nil {
		return fmt.Errorf("Crash while inserting movie!\nerr.Error(): %v\n", err.Error())
	}
	return nil
}

func (movie *Movies) Update(db *sqlx.DB) error {
	_, err := db.Exec(`CALL sp_update_movie($1,$2,$3,$4,$5)`,
		movie.Id,
		movie.Name,
		movie.Description,
		movie.PosterUrl,
		movie.Stars)

	if err != nil {
		return fmt.Errorf("Crash while updating movie!\nerr.Error(): %v\n", err.Error())
	}
	return nil
}

type Series struct {
	Id          int            `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Stars       int16          `db:"stars"`
	PosterId    sql.NullInt32  `db:"poster_id"`
	AddedBy     int            `db:"added_by"`
}

type SeriesPoster struct {
	Id         int            `db:"id"`
	SeriesName sql.NullString `db:"series_name"`
    PosterUrl  sql.NullString `db:"poster_url"`
}
