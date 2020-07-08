package sql

type ConstraintType int

const (
	PrimaryKey ConstraintType = iota + 1
	ForeignKey
	NotNull
	Unique
)

type ConstraintError struct {
	Type ConstraintType

	Cause error
}

func (ce ConstraintError) Error() string {
	return ce.Cause.Error()
}
