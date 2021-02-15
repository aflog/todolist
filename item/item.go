package item

import (
	"errors"
	"time"
)

// Item defines the structure of an to do list task
type Item struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Labels      []Label   `json:"labels"`
	Comments    []Comment `json:"comments"`
	Status      bool      `json:"status"`
	DueDate     time.Time `json:"dueDate"`
	UpdatedAt   time.Time `json:"-"`
	CreatedAt   time.Time `json:"-"`
}

// Label defines the structure of a label used in to do list tasks
type Label struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"-"`
	Text      string    `json:"text"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}

// Comment defines the structure of a comment used in to do list tasks
type Comment struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"-"`
	Text      string    `json:"text"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}

//Validate that all required field are present
func (i *Item) Validate() error {
	if i.Title == "" {
		return errors.New("item: title field is required and can not be empty")
	}
	return nil
}
