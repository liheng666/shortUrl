package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Db struct {
	Username string
	Password string
	Address  string
	Dbname   string
}

func (d *Db) Conn() *sql.DB {
	//dsn := "root:123456@tcp(localhost:3306)/sqlx_db?charset=utf8mb4"
	dsn := fmt.Sprintf("%s:%s@%s/%s", d.Username, d.Password, d.Address, d.Dbname)
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 判断数据库连接是否成功
	err = DB.Ping()
	if err != nil {
		panic(err)
	}

	return DB
}

// 创建数据库
func CreateTables(int) {

}
