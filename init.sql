create table if not exists users (
  id serial primary key,
  username VARCHAR(75) not null,
  email VARCHAR(100) not null unique,
  pfp_url VARCHAR(255),
  password INT not null
);

CREATE INDEX idx_email ON users(email);
CREATE INDEX idx_login ON users(username, password);

create OR REPLACE procedure sp_insert_user(
  OUT v_id INT,
  v_username VARCHAR(75),
  v_email VARCHAR(100),
  v_pfp_url VARCHAR(255),
  v_password INT
)
language plpgsql
as $$
begin
  IF EXISTS (SELECT 1 FROM users WHERE username = v_username) THEN
    RAISE EXCEPTION 'Username already taken!';
  END IF;

  INSERT INTO users(username, email, pfp_url, password)
  VALUES(v_username, v_email, v_pfp_url, v_password)
  RETURNING id INTO v_id;

  EXCEPTION
      WHEN OTHERS THEN
          RAISE;
end;
$$

create or replace procedure sp_update_user(
  v_id int,
  v_username VARCHAR(75),
  v_email VARCHAR(100),
  v_pfp_url VARCHAR(255),
  v_password INT
)
language plpgsql
as $$
begin
  IF EXISTS (SELECT 1 FROM users WHERE username = v_username) THEN
    RAISE EXCEPTION 'Username already taken!';
  END IF;

  update users
  set 
    username = v_username,
    email = v_email,
    pfp_url = v_pfp_url,
    password = v_password
  where
    id = v_id;

  EXCEPTION
      WHEN OTHERS THEN
          RAISE;
end;
$$

create table if not exists movies (
  id SERIAL primary key,
  name VARCHAR(75) not null,
  description VARCHAR(150),
  poster_url VARCHAR(255),
  stars SMALLINT not null default 0 check (stars <= 5),
  added_by INT,
  foreign KEY (added_by) references users (id)
);

CREATE INDEX idx_movie_name ON movies(name);

CREATE OR REPLACE FUNCTION fn_insert_movie(
    v_name VARCHAR(75),
    v_description VARCHAR(150),
    v_poster_url VARCHAR(255),
    v_stars SMALLINT,
    v_added_by INT
) 
RETURNS INT AS $$
DECLARE v_id INT;
BEGIN
  IF EXISTS (SELECT 1 FROM movies WHERE name = v_name) THEN
    RAISE EXCEPTION 'Already exists a movie with that name!';
  END IF;
  
  INSERT INTO movies (name, description, poster_url, stars, added_by)
  VALUES (v_name, v_description, v_poster_url, v_stars, v_added_by)
  RETURNING id INTO v_id;

  RETURN v_id;
  EXCEPTION
      WHEN OTHERS THEN
          RAISE;
END;
$$ LANGUAGE plpgsql;

-- craeted by chatgpt
CREATE OR REPLACE PROCEDURE sp_update_movie(
    p_id INT,
    p_name VARCHAR(75),
    p_description VARCHAR(150),
    p_poster_url VARCHAR(255),
    p_stars SMALLINT
) AS $$
BEGIN
    UPDATE movies
    SET name = p_name,
        description = p_description,
        poster_url = p_poster_url,
        stars = p_stars
    WHERE id = p_id;
END;
$$ LANGUAGE plpgsql;

create table if not exists episodes (
  id SERIAL primary key,
  name VARCHAR(75) not null,
  description VARCHAR(125),
  stars SMALLINT not null default 0 check (stars <= 5),
  poster_id INT,
  added_by INT,
  foreign KEY (added_by) references users (id),
  foreign KEY (poster_id) references series_posters (id)
);

create table if not exists series_posters (
  id SERIAL primary key,
  series_name VARCHAR(50) unique,
  poster_url VARCHAR(255)
);

CREATE INDEX idx_episode_name ON episodes(name);
CREATE INDEX idx_series_name ON series_posters(series_name);

CREATE OR REPLACE FUNCTION fn_insert_episode(
  v_name VARCHAR(75),
  v_description VARCHAR(125),
  v_stars SMALLINT,
  v_poster_id INT,
  v_added_by INT,
  v_series_name VARCHAR(50),
  v_poster_url VARCHAR(255)
) RETURNS INT AS $$
DECLARE v_poster_id INT;
DECLARE v_episode_id INT;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM series_posters WHERE series_name = v_series_name) THEN
        INSERT INTO series_posters(series_name, poster_url)
        VALUES(v_series_name, v_poster_url);
    END IF;

    SELECT id
    INTO v_poster_id
    FROM series_posters
    WHERE series_name = v_series_name;

    INSERT INTO episodes(name, description, stars, poster_id, added_by)
    VALUES(v_name, v_description, v_stars, v_poster_id, v_added_by)
    RETURNING id INTO v_episode_id;
    
    RETURN v_episode_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE PROCEDURE sp_update_episode (
  v_episode_id INT,
  v_name VARCHAR(75),
  v_description VARCHAR(125),
  v_stars SMALLINT,
  v_added_by INT,
  v_series_name VARCHAR(50),
  v_poster_url VARCHAR(255)
) AS $$
DECLARE v_id INT;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM series_posters WHERE series_name = v_series_name) THEN
        INSERT INTO series_posters(series_name, poster_url)
        VALUES(v_series_name, v_poster_url) RETURNING id INTO v_id;
    ELSE    
        SELECT id
        INTO v_id
        FROM series_posters
        WHERE series_name = v_series_name
    END IF;
 
    UPDATE episodes
    SET name = v_name,
    description = v_description,
    stars = v_stars,
    poster_id = v_id,
    added_by = v_added_by
    WHERE id = v_episode_id;
END;
$$ LANGUAGE plpgsql;

