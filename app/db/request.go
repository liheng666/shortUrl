package db

import (
	"time"
	"fmt"
	"database/sql"
)

// 需要储存的到数据库的数据
type Request struct {
	Uid       uint64
	Shortcode string
	UrlStr    string
	Time      time.Time
}

// 插入数据库
func (r *Request) Insert(db *sql.DB) error {

	sql := fmt.Sprintf("insert into short_%d(uid,shortcode,urlstr,time) values(?,?,?,?)", r.Uid%uint64(tableCount))
	stmt, err := db.Prepare(sql)
	defer stmt.Close()

	if err != nil {
		return err
	}
	_, err = stmt.Exec(r.Uid, r.Shortcode, r.UrlStr, r.Time)
	if err != nil {
		return err
	}
	return nil
}

// 查询数据
func (r *Request) Select(db *sql.DB) error {
	sql := fmt.Sprintf("select urlstr from short_%d where uid=?", r.Uid%uint64(tableCount))
	stmt, err := db.Prepare(sql)
	defer stmt.Close()

	if err != nil {
		return err
	}
	row := stmt.QueryRow(r.Uid)
	if err != nil {
		return err
	}
	err = row.Scan(&r.UrlStr)
	if err != nil {
		return err
	}
	return nil
}
