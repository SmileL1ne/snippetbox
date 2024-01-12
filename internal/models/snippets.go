package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title, content string, expires int) (int, error)
	Get(id int) (Snippet, error)
	Latest() ([]Snippet, error)
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (sm *SnippetModel) Insert(title, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets(title, content, created, expires) 
	VALUES(?, ?, DATE_ADD(UTC_TIMESTAMP(), INTERVAL 6 HOUR), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	res, err := sm.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (sm *SnippetModel) Get(id int) (Snippet, error) {
	var snippet Snippet

	stmt := `SELECT * FROM snippets WHERE expires > UTC_TIMESTAMP() AND id = ?`
	if err := sm.DB.QueryRow(stmt, id).Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	return snippet, nil
}

func (sm *SnippetModel) Latest() ([]Snippet, error) {

	stmt := `SELECT * FROM snippets WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`
	rows, err := sm.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		if err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires); err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
