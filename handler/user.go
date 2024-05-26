package handler

import (
	"filestore-serve/db"
	"filestore-serve/util"
	"fmt"
	"net/http"
	"time"
)

const (
	pwd_salt = "*@(s" // 加密盐
)

// SignupHandler 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("The request's method is not post"))
		return
	}
	r.ParseMultipartForm(32 << 20)
	username := r.Form.Get("username")
	pwd := r.Form.Get("password")

	if len(username) < 3 || len(pwd) < 8 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	// 给密码加密
	encPwd := util.Sha1([]byte(pwd + pwd_salt))
	suc := db.UserSignup(username, encPwd)

	if suc {
		w.Write([]byte("Success"))

	} else {
		w.Write([]byte("Fail"))
	}
	return
}

// SignInHandler 用户登录
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("The request's method is not post"))
		return
	}
	r.ParseMultipartForm(32 << 20)

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	encPwd := util.Sha1([]byte(password + pwd_salt))
	// 1. 校验用户名及密码
	pwdChecked := db.UserSignin(username, encPwd)
	if !pwdChecked {
		w.Write([]byte("Fail"))
		return
	}
	// 2. 生成访问凭证（token）
	token := GenToken(username)
	ok := db.UpdateToken(username, token)
	if !ok {
		w.Write([]byte("Filed"))
		return
	}
	// 3. 登录成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))

	//现在我们使用了自己写的一个工具类封装了我们的json操作，并且由于返回的数据量比较大，所以我们推荐使用json作为返回的数据类型
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		//因为我们登录成功要返回让用户重定向页面的url地址，并且还要给用户一个token用于进行其他api的凭证访问
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			// 我们是前后端分离的，这个location实际上在项目中是没有用到的
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	//最后我们将response转换成jsonBytes并返回给客户端
	w.Write(resp.JSONBytes())

}

// GenToken 生成token
func GenToken(username string) string {
	// 40位字符：md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析参数
	r.ParseMultipartForm(32 << 20)
	username := r.Form.Get("username")
	//token := r.Form.Get("token")

	//2. 验证token是否有效
	//isValidToken := ISTokenValid(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	//3. 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//4. 组装并且相应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "ok",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// ISTokenValid 验证token是否有效
func ISTokenValid(token string) bool {
	// 暂且默认为true
	//TODO:判断token的时效性，是否过期
	//TODO：从数据库表tbl_user_token查询username对应的token信息
	//TODO：对比两个token是否一致
	return true
}
