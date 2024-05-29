package handler

import (
	"encoding/json"
	"filestore-serve/db"
	"filestore-serve/meta"
	"filestore-serve/store/ceph"
	"filestore-serve/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// UploadHandler : 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html页面
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "Internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Failed to get data, err:", err.Error())
			return
		}
		// 关闭文件句柄
		defer file.Close()
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "./tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 创建文件
		// 我是在Windows开发的，所以用了这样的路径，Linux直接就是/tmp/
		folderPath := "./tmp/"
		// Windows得先创建文件夹
		os.Mkdir(folderPath, 0755)
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			return
		}
		defer newFile.Close()

		// 将上传的文件内容拷贝到创建的文件中去
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to save data into file, err:", err.Error())
			return
		}

		// 复制是靠移动指针不断读写的，现在将文件指针移回到开头
		newFile.Seek(0, 0)
		// 从头开始计算 sha1 值
		fileMeta.FileSha1 = util.FileSha1(newFile)
		fmt.Println(fileMeta.FileSha1)

		// 同时将文件写入ceph存储
		client := ceph.GetCephConn()
		key := "/ceph/" + fileMeta.FileSha1
		ceph.UploadFile(client, "userfile", key, fileMeta.Location)
		fileMeta.Location = key

		// 保存到fileMetas中
		//meta.UpdateFileMeta(fileMeta)
		// 保存到MySQL中
		ok := meta.UpdateFileMetaDB(fileMeta)
		if !ok {
			w.Write([]byte("File was uploaded"))
			return
		}

		//更新用户文件表记录
		r.ParseMultipartForm(32 << 20)
		username := r.Form.Get("username")
		suc := db.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
	}
}

// UploadSuHandler : 上传已完成
// 应该是解耦操作
func UploadSuHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler 批量获取文件信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas := meta.GetLastFileMetas(limitCnt)
	userfile, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//返回userfile的json数据
	data, err := json.Marshal(userfile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler 文件下载
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//理论上工作是已经做完了的，但是为了让浏览器做一个演示，我们需要将一个http的响应头，让浏览器识别出来就可以当成一个文件进行下载
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler 注意，这里只是修改了元信息的数据，并没有真正修改文件的文件名
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	// apifox 的Body 类型 : multipart/form-data，不能用ParseForm
	r.ParseMultipartForm(32 << 20)

	// 操作operate
	opType := r.PostForm.Get("op")
	fileSha1 := r.PostForm.Get("filesha")
	newFileName := r.PostForm.Get("filename")

	// 这里仅支持重命名文件，操作数为0
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	fmt.Println(curFileMeta)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	// apifox 的Body 类型 : multipart/form-data，不能用ParseForm
	r.ParseMultipartForm(32 << 20)

	fileSha1 := r.PostForm.Get("filesha")

	fMeta := meta.GetFileMeta(fileSha1)
	err := os.Remove(fMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler 秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")

	// 2. 从文件表中查询相同hash的文件记录
	fMeta, err := meta.GetFileMetaDB(filehash)
	// 找不到就直接报错了
	if err != nil {
		// 3. 查不到记录则返回秒数失败
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表，返回成功
	// 没必要用户上传文件大小，获取到有数据就直接用它的，因为文件内容不变，hash就不变，大小也不变
	// 文件名要用户上传，因为可能文件名不同，但是内容相同
	suc := db.OnUserFileUploadFinished(username, filehash, filename, fMeta.FileSize)
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}
