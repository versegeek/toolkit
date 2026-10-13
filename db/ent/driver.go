package ent

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"entgo.io/ent/dialect"
	// Register the database/sql drivers used by the built-in backends.
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/versegeek/toolkit/internal/sqlconn"
)

// driverSpec names the database/sql driver and the ent dialect of a backend.
type driverSpec struct {
	sqlDriver string
	dialect   string
}

var (
	driversMu sync.RWMutex
	drivers   = map[string]driverSpec{
		sqlconn.DriverMySQL:    {sqlDriver: "mysql", dialect: dialect.MySQL},
		sqlconn.DriverPostgres: {sqlDriver: "pgx", dialect: dialect.Postgres},
		sqlconn.DriverSQLite:   {sqlDriver: "sqlite", dialect: dialect.SQLite},
	}
)

// RegisterDriver makes a backend available under name for Config.Driver.
// sqlDriver is the name registered with database/sql and entDialect one of
// the entgo.io/ent/dialect constants, for example:
//
//	ent.RegisterDriver("tidb", "mysql", dialect.MySQL)
//
// Such backends need Config.DSN unless name is a built-in driver name. Call
// it before Init.
func RegisterDriver(name, sqlDriver, entDialect string) {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers[name] = driverSpec{sqlDriver: sqlDriver, dialect: entDialect}
}

func lookupDriver(name string) (driverSpec, error) {
	if name == "" {
		return driverSpec{}, sqlconn.ErrNoDriver
	}
	driversMu.RLock()
	defer driversMu.RUnlock()

	spec, ok := drivers[name]
	if !ok {
		names := make([]string, 0, len(drivers))
		for n := range drivers {
			names = append(names, n)
		}
		slices.Sort(names)
		return driverSpec{}, fmt.Errorf("unsupported driver %q (available: %s)", name, strings.Join(names, ", "))
	}
	return spec, nil
}
