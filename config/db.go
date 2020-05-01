package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// OpenDB opens a DB connection with the mssql server.
func OpenDB() (*gorm.DB, error) {
	// Connect to the database once on startup
	dbConfig := mysql.NewConfig()
	dbConfig.User = "root"
	dbConfig.Passwd = ""
	dbConfig.DBName = "example"
	dbConfig.ParseTime = true
	dbConfig.InterpolateParams = true
	dbConfig.Net = "tcp"
	dbConfig.Addr = fmt.Sprintf("%s:%s", "127.0.0.1", "3306")
	dbConfig.Collation = "utf8_general_ci"

	dsn := dbConfig.FormatDSN()

	mysqlDb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open("mysql", mysqlDb)
	if err != nil {
		return nil, err
	}

	mysql.ErrInvalidConn = driver.ErrBadConn

	// Back of slightly from server configuration which is currently 28800 (8 hours).
	// SHOW VARIABLES WHERE Variable_name = 'wait_timeout';
	db.DB().SetConnMaxLifetime(25 * time.Second)
	db.DB().SetMaxIdleConns(10)
	db.LogMode(true)

	return db, nil
}
