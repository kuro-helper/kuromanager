package erogssync

import (
	_ "embed"
	"strconv"
	"strings"
)

const (
	placeholderMin = "{{MIN_ID}}"
	placeholderMax = "{{MAX_ID}}"
)

//go:embed dump.sql
var dumpSQL string

func buildDumpSQL(minID, maxID int) string {
	sql := strings.ReplaceAll(dumpSQL, placeholderMin, strconv.Itoa(minID))
	return strings.ReplaceAll(sql, placeholderMax, strconv.Itoa(maxID))
}
