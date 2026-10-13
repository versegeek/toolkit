package gorm

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	gormsqlite "github.com/glebarez/sqlite"
	gormclickhouse "gorm.io/driver/clickhouse"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/versegeek/toolkit/internal/sqlconn"
)

var (
	dialectorsMu sync.RWMutex
	dialectors   = map[string]func(dsn string) gorm.Dialector{
		sqlconn.DriverMySQL:      gormmysql.Open,
		sqlconn.DriverPostgres:   gormpostgres.Open,
		sqlconn.DriverSQLite:     gormsqlite.Open,
		sqlconn.DriverClickHouse: gormclickhouse.Open,
	}
)

// RegisterDialector makes a GORM dialector available under name for
// Config.Driver, for example:
//
//	gorm.RegisterDialector("sqlserver", sqlserver.Open)
//
// Such drivers need Config.DSN. Call it before Init.
func RegisterDialector(name string, open func(dsn string) gorm.Dialector) {
	dialectorsMu.Lock()
	defer dialectorsMu.Unlock()
	dialectors[name] = open
}

func dialector(name string) (func(dsn string) gorm.Dialector, error) {
	if name == "" {
		return nil, sqlconn.ErrNoDriver
	}
	dialectorsMu.RLock()
	defer dialectorsMu.RUnlock()

	open, ok := dialectors[name]
	if !ok {
		names := make([]string, 0, len(dialectors))
		for n := range dialectors {
			names = append(names, n)
		}
		slices.Sort(names)
		return nil, fmt.Errorf("unsupported driver %q (available: %s)", name, strings.Join(names, ", "))
	}
	return open, nil
}
