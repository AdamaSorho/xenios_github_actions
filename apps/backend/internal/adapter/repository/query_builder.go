package repository

import (
	"fmt"
	"strings"
)

// queryBuilder constructs parameterized SQL WHERE clauses.
// All filter values are passed as parameterized arguments ($1, $2, ...) to prevent injection.
type queryBuilder struct {
	conditions []string
	args       []interface{}
	paramIdx   int
}

// newQueryBuilder creates a new queryBuilder starting at parameter index 1.
func newQueryBuilder() *queryBuilder {
	return &queryBuilder{paramIdx: 1}
}

// addCondition adds an equality condition (column = $N).
func (b *queryBuilder) addCondition(column string, value interface{}) {
	b.conditions = append(b.conditions, fmt.Sprintf("%s = $%d", column, b.paramIdx))
	b.args = append(b.args, value)
	b.paramIdx++
}

// addTimeCondition adds a comparison condition (column op $N).
func (b *queryBuilder) addTimeCondition(column, op string, value interface{}) {
	b.conditions = append(b.conditions, fmt.Sprintf("%s %s $%d", column, op, b.paramIdx))
	b.args = append(b.args, value)
	b.paramIdx++
}

// whereClause returns the WHERE clause string (including " WHERE ") or empty string.
func (b *queryBuilder) whereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conditions, " AND ")
}

// nextParamIdx returns the next available parameter index.
func (b *queryBuilder) nextParamIdx() int {
	return b.paramIdx
}
