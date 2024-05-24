package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

var db *sql.DB

func init() {
	user := "root"
	password := "123456"
	ip := "127.0.0.1"
	port := "3306"
	dbName := "filestore"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, password, ip, port, dbName)

	db, _ = sql.Open("mysql", dsn)

	//连接池
	db.SetMaxOpenConns(100)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect mysql, err:", err.Error())
		os.Exit(1)
	}
}

// DBConn 返回数据库连接对象
func DBConn() *sql.DB {
	return db
}
