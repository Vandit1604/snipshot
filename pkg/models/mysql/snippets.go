package mysql

import (
	"database/sql"

	"github.com/vandit1604/snipshot/pkg/models"
)

type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	stmt := `INSERT INTO snippets 
	(title, content, created, expires)
	VALUES
	(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt, err := m.DB.Prepare(`SELECT id, title, content,created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() AND id=?`)
	if err != nil {
		return nil, err
	}
	rows := stmt.QueryRow(id)

	s := models.Snippet{}
	err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrRecordNotFound
	}

	return &s, nil
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt, err := m.DB.Prepare(`SELECT id, title, content,created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*models.Snippet

	for rows.Next() {
		var s models.Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		if rows.Err() != nil {
			return nil, err
		}

		snippets = append(snippets, &s)
	}

	return snippets, nil
}
