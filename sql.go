package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	// 1997-08-13T04:00:00.000Z
	// create a go time layout from the above
	ISO8601 = "2006-01-02 15:04:05.000"

	createTable = `
CREATE TABLE IF NOT EXISTS southpark (
	season INTEGER,
	episode INTEGER,
	title TEXT,
	url TEXT,
	description TEXT,
	imageUrl TEXT,
	date TEXT,
	PRIMARY KEY (season, episode)
);

CREATE INDEX IF NOT EXISTS idx_southpark_url ON southpark (url);
`

	insertVideo = `
INSERT OR REPLACE INTO southpark (
	season, 
	episode, 
	title, 
	url, 
	description, 
	imageUrl,
	date
	) VALUES (?, ?, ?, ?, ?, ?, ?);
`

	lastUrl = `
SELECT url FROM southpark
ORDER BY season DESC, episode DESC
LIMIT 1;
`

	visitedUrl = `
SELECT url FROM southpark WHERE url = ?;
`

	seasonVideos = `
SELECT season, episode, title, url, description, imageUrl, date FROM southpark WHERE season = ?;
	`

	episodeVideo = `
SELECT season, episode, title, url, description, imageUrl, date FROM southpark WHERE season = ? AND episode = ?;
	`

	allVideos = `
SELECT season, episode, title, url, description, imageUrl, date FROM southpark;
`
)

var ErrNotFound = errors.New("no entries found")

func (c *rootContext) InitDB() error {
	db, err := sql.Open("sqlite", c.Config.DBPath())
	if err != nil {
		return err
	}

	c.DB = db

	_, err = c.DB.ExecContext(c.Ctx, createTable)
	if err != nil {
		return err
	}
	return nil
}

func (c *rootContext) CloseDB() error {
	return c.DB.Close()
}

func (c *rootContext) Insert(title string, season, episode int, url, description, imageUrl string, contentDate time.Time) error {

	_, err := c.DB.ExecContext(c.Ctx, insertVideo, season, episode, title, url, description, imageUrl, contentDate.Format(ISO8601))
	if err != nil {
		return err
	}
	return nil
}

func (c *rootContext) Last() (string, error) {
	rows, err := c.DB.QueryContext(c.Ctx, lastUrl)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			return "", err
		}
		urls = append(urls, url)
	}

	if len(urls) == 0 {
		return "", ErrNotFound
	}

	return urls[0], nil
}

func (c *rootContext) Visited(url string) (bool, error) {
	row := c.DB.QueryRowContext(c.Ctx, visitedUrl, url)

	var u string
	err := row.Scan(&u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

type Video struct {
	Title       string
	Season      int
	Episode     int
	Url         string
	Description string
	ImageUrl    string
	Date        time.Time
}

func (v *Video) Format() string {
	return fmt.Sprintf("South_Park_S%02dE%02d.%%(ext)s", v.Season, v.Episode)
}

func (v *Video) SeasonString() string {
	return fmt.Sprintf("S%02d", v.Season)
}

func (c *rootContext) Season(season int) ([]Video, error) {
	rows, err := c.DB.QueryContext(c.Ctx, seasonVideos, season)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		date := ""
		err := rows.Scan(&v.Season, &v.Episode, &v.Title, &v.Url, &v.Description, &v.ImageUrl, &date)
		if err != nil {
			return nil, err
		}

		t, err := time.Parse(ISO8601, date)
		if err != nil {
			return nil, err
		}
		v.Date = t
		videos = append(videos, v)
	}

	if len(videos) == 0 {
		return nil, ErrNotFound
	}

	return videos, nil
}

func (c *rootContext) Episode(season, episode int) (video Video, err error) {
	row := c.DB.QueryRowContext(c.Ctx, episodeVideo, season, episode)

	var v Video
	date := ""
	err = row.Scan(&v.Season, &v.Episode, &v.Title, &v.Url, &v.Description, &v.ImageUrl, &date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return video, ErrNotFound
		}
		return video, err
	}

	t, err := time.Parse(ISO8601, date)
	if err != nil {
		return video, err
	}
	v.Date = t

	return v, nil
}

func (c *rootContext) All() ([]Video, error) {
	rows, err := c.DB.QueryContext(c.Ctx, allVideos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		date := ""
		err := rows.Scan(&v.Season, &v.Episode, &v.Title, &v.Url, &v.Description, &v.ImageUrl, &date)
		if err != nil {
			return nil, err
		}

		t, err := time.Parse(ISO8601, date)
		if err != nil {
			return nil, err
		}
		v.Date = t
		videos = append(videos, v)
	}

	if len(videos) == 0 {
		return nil, ErrNotFound
	}

	return videos, nil
}
