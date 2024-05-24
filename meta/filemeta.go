package meta

import (
	"filestore-serve/db"
	"sort"
)

// FileMeta : 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

///////////////////////////////////////////////////////////////

// UpdateFileMeta 新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// GetFileMeta 通过sha1值获取文件元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetLastFileMetas 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	sort.Sort(ByUploadTime(fMetaArray))

	//为了避免用户输入的数量大于我们总文件数量，避免panic
	if len(fMetaArray) >= count {
		return fMetaArray[0:count]
	} else {
		return fMetaArray[0:len(fMetaArray)]
	}
}

// RemoveFileMeta 	删除文件元信息
// 生产环境中我们需要做一些安全的判断，比如delete操作会不会引起线程同步的问题，如果多线程必须保证map安全
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

//////////////////////////////////////////////////////////////

// UpdateFileMetaDB 新增/更新文件元信息保存到MySQL中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return db.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// GetFileMetaDB 在MySQL通过sha1值获取文件元信息对象
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := db.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	return FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}, nil
}
