package sqlparser

import (
	"strings"
	"unicode"
)

type StmtType uint8

const (
	StmtSelect StmtType = iota
	StmtInsert
	StmtUpdate
	StmtDelete
	StmtUnknown
)

type SQLParser interface {
	GetStatementType(string) StmtType
}

func IsDML(t StmtType) bool {
	if t == StmtSelect {
		return false
	}

	return true
}

func DefaultSQLParser() SQLParser { return sqlParser{} }

type sqlParser struct{}

func (sqlParser) GetStatementType(stmt string) StmtType {
	firstWord := strings.TrimLeftFunc(
		stmt,
		func(r rune) bool { return !unicode.IsLetter(r) },
	)

	if end := strings.IndexFunc(firstWord, unicode.IsSpace); end != -1 {
		firstWord = firstWord[:end]
	}

	switch strings.ToLower(firstWord) {
	case "select":
		return StmtSelect
	case "insert":
		return StmtInsert
	case "update":
		return StmtUpdate
	case "delete":
		return StmtDelete
	}

	return StmtUnknown
}
