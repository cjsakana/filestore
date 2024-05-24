package db

import (
	"database/sql"
	"filestore-serve/db/mysql"
	"fmt"
)

// OnFileUploadFinished 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	// Prepare 预编译，防止sql注入攻击
	stmt, err := mysql.DBConn().Prepare(
		// gland报错不要紧，因为是字符串拼接用了+加号，而goland认为这是SQL，不应该有+加号
		"insert ignore into tbl_file(`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values(?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:", err.Error())
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	// SQL是执行成功了，假如原来就有该文件，filesha1就重复了
	// 执行成功是不会报错的，但是还可以查看受影响的条数
	rf, err := res.RowsAffected()
	if err == nil {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
			return false
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta 从MySQL获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mysql.DBConn().Prepare(
		"select file_size,file_name,file_addr,file_sha1 from tbl_file where file_sha1 = ? and status = 1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(&tfile.FileSize, &tfile.FileName, &tfile.FileAddr, &tfile.FileHash)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tfile, nil
}
