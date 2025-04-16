package renderer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/jmoiron/sqlx"
)

type Users struct {
	Id       int            `db:"id"`
	Username string         `db:"username"`
    Email    string         `db:"email" json:"-"`
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

func (user Users) ToJson() ([]byte, error) {
    userJson, err := json.Marshal(user)    
    if err != nil {
        return nil, fmt.Errorf("Crash while parsing the user to json!\nerr.Error(): %v\n", err.Error())
    }
    return userJson, nil
}

func (user *Users) FromJson(userJson[]byte) error {
    err := json.Unmarshal(userJson, user)
    if err != nil {
        return fmt.Errorf("Crash while serializing user from json!\nerr.Error() %v\n", err.Error())
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
	err := db.QueryRowx(`SELECT * FROM fn_insert_movie($1,$2,$3,$4,$5)`,
		movie.Name,
		movie.Description,
		movie.PosterUrl,
		movie.Stars,
		movie.AddedBy).Scan(&movie.Id)

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

type Episode struct {
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

type SeriesEpisode struct {
    Serie Episode
    Poster SeriesPoster
}

func (episode *SeriesEpisode) Insert(db *sqlx.DB) error {
    err := db.QueryRowx(`SELECT * FROM fn_insert_episode($1,$2,$3,$4,$5,$6,$7)`,
        episode.Serie.Name,
        episode.Serie.Description,
        episode.Serie.Stars,
        episode.Serie.PosterId,
        episode.Serie.AddedBy,
        episode.Poster.SeriesName,
        episode.Poster.PosterUrl).Scan(&episode.Serie.Id)
    if err != nil {
        return fmt.Errorf("Crash while inserting the episode!\nerr.Error(): %v\n", err.Error())
    }
    return nil
}

func (episode *SeriesEpisode) Update(db *sqlx.DB) error {
    _, err := db.Exec(`CALL sp_update_episode($1,$2,$3,$4,$5,$6)`,
        episode.Serie.Id,
        episode.Serie.Name,
        episode.Serie.Description,
        episode.Serie.Stars,
        episode.Poster.SeriesName,
        episode.Poster.PosterUrl)
    if err != nil {
        return fmt.Errorf("Crash while inserting the episode!\nerr.Error(): %v\n", err.Error())
    }
    return nil
}

func GetSeriesPosters(db *sqlx.DB) ([]SeriesPoster, error) {
    var seriesPosters []SeriesPoster
    err := db.Select(&seriesPosters, `SELECT * FROM series_posters WHERE series_name != $1`, DEFAULT_SERIES_POSTER_NAME)
    if err != nil {
        return nil, fmt.Errorf("Crash while getting the series posters!\nerr.Error(): %v\n", err.Error())
    }
    seriePoster := SeriesPoster{
        Id: 0,
        SeriesName: sql.NullString{ String: DEFAULT_SERIES_POSTER_NAME, Valid: true },
        PosterUrl: sql.NullString{ String: DEFAULT_SERIES_POSTER, Valid: true },
    }
    seriesPosters = append(seriesPosters, seriePoster)
    slices.Reverse(seriesPosters)

    return seriesPosters, nil
}

func FetchMovieList(db *sqlx.DB) ([]Movies, error) {
    query := `
        SELECT *
        FROM movies`

    var movieList []Movies
    err := db.Select(&movieList, query)
    if err != nil {
        return nil, fmt.Errorf("Crash while fetching the movie list!\nerr.Error(): %v\n", err.Error())
    }
    return movieList, nil
}

func convertPosterListToMap(series []SeriesPoster) (map[int]SeriesPoster) {
    if len(series) == 0 {
        return nil
    }

    serieMap := map[int]SeriesPoster{}
    for _,serie := range series {
        serieMap[serie.Id] = serie
    }
    return serieMap
}

func FetchEpisodesList(db *sqlx.DB) ([]SeriesEpisode, error) {
    query := `
        SELECT *
        FROM episodes`
    var seriesEpisodes []SeriesEpisode
    var episodes []Episode
    err := db.Select(&episodes, query)
    if err != nil {
        return nil, fmt.Errorf("Crash while fetching episodes list!\nerr.Error(): %v\n", err.Error())
    }

    query = `
        SELECT *
        FROM series_posters`
    var posters []SeriesPoster
    err = db.Select(&posters, query)
    if err != nil {
        return nil, fmt.Errorf("Crash while fetching poster list!\nerr.Error(): %v\n", err.Error())
    }
    posterMap := convertPosterListToMap(posters)
    if posterMap == nil {
        return nil, fmt.Errorf("Crash while converting to poster list to map\n")
    }

    var seriesEpisode SeriesEpisode
    for _, episode := range episodes {
        seriesEpisode = SeriesEpisode{Serie: episode, Poster: posterMap[int(episode.PosterId.Int32)]}
        fmt.Println(seriesEpisode)
        seriesEpisodes = append(seriesEpisodes, seriesEpisode)
    }

    return seriesEpisodes, nil
}

