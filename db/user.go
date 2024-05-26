package db

import (
	"filestore-serve/db/mysql"
	"fmt"
	"log"
)

// UserSignup 用户注册
func UserSignup(username string, password string) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err:", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)

	if err != nil {
		fmt.Println("Failed to insert, err:", err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); err == nil {
		if rowsAffected <= 0 {
			return false
		}
	}
	return true
}

// UserSignin 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mysql.DBConn().Prepare("select * from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("Username no found:" + username)
		return false
	}
	defer rows.Close()

	// 创建一个map来存储查询结果
	var id int
	var userName string
	var userPwd string
	var email string
	var phone string
	var emailValidated int
	var phoneValidated int
	var signupAt string
	var lastActive string
	var profile string
	var status int

	// 遍历结果集
	if rows.Next() {
		// 扫描每一列到变量中，真的麻烦，因为查询是所有*，所以Scan也必须是所有字段都存在
		// 直接引用map还不行，因为mapKV的地址不是固定的，所以不能被引用
		err := rows.Scan(&id, &userName, &userPwd, &email, &phone, &emailValidated, &phoneValidated,
			&signupAt, &lastActive, &profile, &status)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	if id > 0 && userPwd == encpwd {
		return true
	}
	return false
}

// UpdateToken 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := mysql.DBConn().Prepare(
		"replace into tbl_user_token(`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 定义一个用户结构体，与Mysql中user表的结构一一对应
type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mysql.DBConn().Prepare(
		"select user_name, signup_at from tbl_user where user_name = ? limit 1")
	if err != nil {
		log.Println(err.Error())
		return user, err
	}

	//即使关闭资源
	defer stmt.Close()

	//执行查询操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		log.Println(err.Error())
		return user, err
	}
	return user, nil
}
