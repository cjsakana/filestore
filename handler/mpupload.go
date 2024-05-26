package handler

import (
	"context"
	"filestore-serve/cache/redis"
	"filestore-serve/db"
	"filestore-serve/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// 分块初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitialMultipartUploadHandler 初始化分块信息
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseMultipartForm(32 << 20)
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 获得redis的连接
	rConn := redis.NewRedisClient()
	defer rConn.Close()

	// 3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 4. 将初始化信息写入到redis缓存
	ctx := context.Background()
	rConn.HSet(ctx, "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.HSet(ctx, "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.HSet(ctx, "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	// 5. 将响应初始化数据返回到客户端

	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// UploadPartHandler 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseMultipartForm(32 << 20)
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunckIndex, _ := strconv.Atoi(r.Form.Get("index"))

	// 2. 获得redis的连接
	rConn := redis.NewRedisClient()
	defer rConn.Close()

	// 3. 获取文件句柄，用于存储分块内容
	folderPath := "./data/"
	// Windows得先创建文件夹
	os.Mkdir(folderPath, 0755)
	fd, err := os.Create("./data/" + uploadID + "/" + strconv.Itoa(chunckIndex))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	// 没有校验每一块数据的sha1，是否匹配，可能有人篡改数据
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态
	ctx := context.Background()
	// 每一块都记录一下
	rConn.HSet(ctx, "MP_"+uploadID, "chkidx_"+strconv.Itoa(chunckIndex), 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseMultipartForm(32 << 20)
	username := r.Form.Get("username")
	upid := r.Form.Get("uploadid")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2. 获得redis的连接
	rConn := redis.NewRedisClient()
	defer rConn.Close()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.NewRedisClient().HGetAll(context.Background(), "MP_"+upid).Result()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	// 总的有多少
	totalCount := 0
	// 记录当前有多少
	chunkCount := 0
	for k, v := range data {
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount += 1
		}
	}
	if totalCount == chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// 4. TODO：合并分块

	// 5. 更新唯一文件表及用户文件表
	fsize, _ := strconv.ParseInt(filesize, 10, 64)
	db.OnFileUploadFinished(filehash, filename, fsize, "")
	db.OnUserFileUploadFinished(username, filehash, filename, fsize)

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}
