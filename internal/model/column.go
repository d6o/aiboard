package model

// Column represents a kanban board column.
type Column string

const (
	ColumnTodo  Column = "todo"
	ColumnDoing Column = "doing"
	ColumnDone  Column = "done"
)

func (c Column) Valid() bool {
	return c == ColumnTodo || c == ColumnDoing || c == ColumnDone
}
