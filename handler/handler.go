package handler

import (
	"encoding/json"
	"filestore-serve/meta"
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
		// 保存到fileMetas中
		meta.UpdateFileMeta(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
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
	fMeta := meta.GetFileMeta(filehash)
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	fileMetas := meta.GetLastFileMetas(limitCnt)

	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

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
