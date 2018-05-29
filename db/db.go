package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type Db struct {
	Username string
	Password string
	Address  string
	Dbname   string
}

func (d *Db) Conn() *sql.DB {
	//dsn := "root:123456@tcp(localhost:3306)/sqlx_db?charset=utf8mb4"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", d.Username, d.Password, d.Address, d.Dbname)

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

// 创建数据库表
func CreateTables(DB *sql.DB, n int) {

	str := "CREATE TABLE IF NOT EXISTS %s (" +
		"uid BIGINT UNSIGNED," +
		"shortcode VARCHAR(20) NOT NULL," +
		"urlstr VARCHAR(500) NOT NULL," +
		"time DATETIME," +
		"INDEX (uid)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8"
	for i := 0; i < n; i++ {
		sql := fmt.Sprintf(str, "short_"+strconv.Itoa(i))

		_, err := DB.Exec(sql)
		if err != nil {
			panic(err)
		}
	}
}

