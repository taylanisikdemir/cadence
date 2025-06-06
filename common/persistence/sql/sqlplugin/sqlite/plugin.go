// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sqlite

import (
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"

	"github.com/uber/cadence/common/config"
	pt "github.com/uber/cadence/common/persistence/persistence-tests"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/sql/sqldriver"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

const (
	PluginName = "sqlite"
)

// SQLite plugin provides an sql persistence storage implementation for sqlite database
// Mostly the implementation reuses the mysql implementation
// If DatabaseName is not provided, then sqlite will use in-memory database,
// otherwise it will use the file as the database
type plugin struct{}

var _ sqlplugin.Plugin = (*plugin)(nil)

func init() {
	sql.RegisterPlugin(PluginName, &plugin{})
}

// CreateDB wraps createDB to return an instance of sqlplugin.DB
func (p *plugin) CreateDB(cfg *config.SQL) (sqlplugin.DB, error) {
	return p.createDB(cfg)
}

// CreateAdminDB wraps createDB to return an instance of sqlplugin.AdminDB
func (p *plugin) CreateAdminDB(cfg *config.SQL) (sqlplugin.AdminDB, error) {
	return p.createDB(cfg)
}

// createDB create a new instance of DB
func (p *plugin) createDB(cfg *config.SQL) (*DB, error) {
	conns, err := sqldriver.CreateDBConnections(cfg, p.createSingleDBConn)
	if err != nil {
		return nil, err
	}
	return NewDB(conns, nil, sqlplugin.DbShardUndefined, cfg.NumShards, newConverter(), cfg.DatabaseName)
}

// createSingleDBConn creates a single database connection for sqlite
// Plugin respects the following arguments MaxConns, MaxIdleConns, MaxConnLifetime
// Other arguments are used and described in buildDSN function
func (p *plugin) createSingleDBConn(cfg *config.SQL) (*sqlx.DB, error) {
	if cfg.DatabaseName == "" {
		return p.createDBConn(cfg)
	}

	return createSharedDBConn(cfg.DatabaseName, func() (*sqlx.DB, error) {
		return p.createDBConn(cfg)
	})
}

func (p *plugin) createDBConn(cfg *config.SQL) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", buildDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %v", err)
	}

	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(cfg.MaxConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxConnLifetime)
	}

	// Maps struct names in CamelCase to snake without need for DB struct tags.
	db.MapperFunc(strcase.ToSnake)
	return db, nil
}

const testSchemaDir = "schema/sqlite"

// GetTestClusterOption returns a test cluster option for sqlite plugin
// It uses a temporary directory for the database name
func GetTestClusterOption() *pt.TestBaseOptions {
	return &pt.TestBaseOptions{
		DBPluginName: PluginName,
		DBName:       path.Join(os.TempDir(), uuid.New().String()),
		SchemaDir:    testSchemaDir,
		StoreType:    config.StoreTypeSQL,
	}
}
