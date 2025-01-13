package mock

import (
	"time"

	"github.com/vandit1604/snipshot/pkg/models"
)

// created a mock snippet to return and use in testing
var mockSnippet = &models.Snippet{
	ID:      1,
	Title:   "Ao Kabhi Haveli Pe",
	Content: "Ao Kabhi Haveli Pe...",
	Created: time.Now(),
	Expires: time.Now(),
}

// created an empty SnippetModel struct to create functions against. It's only use is to encapsulate the functions with itself
type SnippetModel struct{}

func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	return 2, nil
}

func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return nil, models.ErrRecordNotFound
	}
}

func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}
