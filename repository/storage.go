package repository

import (
	"context"

	"github.com/aflog/todolist/item"
)

//Repository defines an interface for items storage
type Repository interface {
	CreateItem(ctx context.Context, i item.Item) (int, error)
	GetItems(ctx context.Context) ([]item.Item, error)
	GetItem(ctx context.Context, id int) (*item.Item, error)
}
