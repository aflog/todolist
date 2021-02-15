package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aflog/todolist/item"
	"github.com/go-sql-driver/mysql"
)

// Repository holds the data needed for storing in mysql DB
// It implements the repository.Repository interface
type Repository struct {
	db *sql.DB
}

// NewRepository creates and sets Repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateItem stores provided item and returns its id
// Item, labels and comments are stored in one transaction
func (r *Repository) CreateItem(ctx context.Context, i item.Item) (int, error) {
	// prepare transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	// insert item
	var due *time.Time
	if !i.DueDate.IsZero() {
		due = &i.DueDate
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO item(title, description, status, due) VALUES (?, ?, ?, ?)", i.Title, i.Description, i.Status, due)
	if err != nil {
		tx.Rollback()
		log.Println("item inserting")
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	createdID := int(id)

	// insert comments
	for _, c := range i.Comments {
		_, err := tx.ExecContext(ctx, "INSERT INTO comment(itemId, comment) VALUES (?, ?)", createdID, c.Text)
		if err != nil {
			tx.Rollback()
			log.Println("comment inserting")
			return 0, err
		}
	}

	// insert labels
	for _, l := range i.Labels {
		_, err := tx.ExecContext(ctx, "INSERT INTO label(itemId, label) VALUES (?, ?)", createdID, l.Text)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// execute transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return createdID, err
}

// GetItems returns list of all the items
func (r *Repository) GetItems(ctx context.Context) ([]item.Item, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description, status, due FROM item")
	if err != nil {
		return nil, err
	}

	labels, err := r.getAllLabels(ctx)
	if err != nil {
		return nil, err
	}

	comments, err := r.getAllComments(ctx)
	if err != nil {
		return nil, err
	}

	items := []item.Item{}
	for rows.Next() {
		i := item.Item{}
		dueDate := mysql.NullTime{}
		if err := rows.Scan(&i.ID, &i.Title, &i.Description, &i.Status, &dueDate); err != nil {
			return nil, err
		}
		i.Comments = comments[i.ID]
		i.Labels = labels[i.ID]
		if dueDate.Valid {
			i.DueDate = dueDate.Time
		}
		items = append(items, i)
	}

	return items, nil
}

// GetItem returns an item corresponding to the provided id and nil if it doesn't exist
func (r *Repository) GetItem(ctx context.Context, id int) (*item.Item, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description, status, due FROM item WHERE id=? LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels, err := r.getLabelsByID(ctx, []int{id})
	if err != nil {
		return nil, err
	}

	comments, err := r.getCommentsByID(ctx, []int{id})
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		i := &item.Item{}
		dueDate := mysql.NullTime{}
		if err := rows.Scan(&i.ID, &i.Title, &i.Description, &i.Status, &dueDate); err != nil {
			return nil, err
		}
		i.Comments = comments[i.ID]
		i.Labels = labels[i.ID]
		if dueDate.Valid {
			i.DueDate = dueDate.Time
		}
		return i, nil
	}

	return nil, nil
}

func (r *Repository) getLabelsByID(ctx context.Context, itemIds []int) (map[int][]item.Label, error) {
	itemIdsStr := make([]string, len(itemIds))
	for i, value := range itemIds {
		itemIdsStr[i] = strconv.Itoa(value)
	}
	sqlStatement := fmt.Sprintf("SELECT id, itemId, label FROM label WHERE itemId IN(%s)", strings.Join(itemIdsStr, ", "))
	return r.getLabels(ctx, sqlStatement)
}

func (r *Repository) getAllLabels(ctx context.Context) (map[int][]item.Label, error) {
	sqlStatement := "SELECT id, itemId, label FROM label"
	return r.getLabels(ctx, sqlStatement)
}

func (r *Repository) getLabels(ctx context.Context, sql string) (map[int][]item.Label, error) {
	rows, err := r.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	labels := make(map[int][]item.Label)
	for rows.Next() {
		i := item.Label{}
		if err := rows.Scan(&i.ID, &i.ItemID, &i.Text); err != nil {
			return nil, err
		}
		labels[i.ItemID] = append(labels[i.ItemID], i)
	}

	return labels, nil
}

func (r *Repository) getCommentsByID(ctx context.Context, itemIds []int) (map[int][]item.Comment, error) {
	itemIdsStr := make([]string, len(itemIds))
	for i, value := range itemIds {
		itemIdsStr[i] = strconv.Itoa(value)
	}
	sqlStatement := fmt.Sprintf("SELECT id, itemId, comment FROM comment WHERE itemId IN(%s)", strings.Join(itemIdsStr, ", "))
	return r.getComments(ctx, sqlStatement)
}

func (r *Repository) getAllComments(ctx context.Context) (map[int][]item.Comment, error) {
	sqlStatement := "SELECT id, itemId, comment FROM comment"
	return r.getComments(ctx, sqlStatement)
}

func (r *Repository) getComments(ctx context.Context, sql string) (map[int][]item.Comment, error) {
	rows, err := r.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	comments := make(map[int][]item.Comment)
	for rows.Next() {
		i := item.Comment{}
		if err := rows.Scan(&i.ID, &i.ItemID, &i.Text); err != nil {
			return nil, err
		}
		comments[i.ItemID] = append(comments[i.ItemID], i)
	}

	return comments, nil
}
