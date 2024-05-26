package db

import (
	"filestore-serve/db/mysql"
	"fmt"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, err := mysql.DBConn().Prepare(
		"insert ignore into tbl_user_file(`user_name`,`file_sha1`,`file_name`, " +
			"`file_size`,`upload_at`) values (?,?,?,?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// QueryUserFileMetas 用户批量获取文件信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mysql.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at,last_update " +
			"from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)

	var userFile []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize,
			&ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		userFile = append(userFile, ufile)
	}
	return userFile, nil
}
