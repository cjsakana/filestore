package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Sha1Stream 用于在流式数据处理时计算数据的 SHA1 哈希值
// SHA1的中文名是：安全散列算法1
type Sha1Stream struct {
	_sha1 hash.Hash
}

// Update 方法用于追加数据到 SHA1 哈希对象中
func (obj *Sha1Stream) Update(data []byte) {
	if obj._sha1 == nil {
		obj._sha1 = sha1.New()
	}
	obj._sha1.Write(data)
}

// Sum 方法用于计算当前数据的 SHA1 哈希值，并以十六进制字符串的形式返回
func (obj *Sha1Stream) Sum() string {
	//SHA1 算法要求输入数据的长度必须是 512 位的整数倍。
	//由于 SHA1 散列结果的长度是 20 字节，所以使用一个空字符串可以确保计算出的散列值是 20 字节长度的。
	return hex.EncodeToString(obj._sha1.Sum([]byte("")))
}

// Sha1 用于计算数据的 SHA1 哈希值
func Sha1(data []byte) string {
	_sha1 := sha1.New()
	_sha1.Write(data)
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}

// FileSha1 用于计算文件的 SHA1 哈希值
func FileSha1(file *os.File) string {
	_sha1 := sha1.New()
	io.Copy(_sha1, file)
	//nil 是一个空切片，意味着没有提供任何数据给 SHA1 哈希计算器。
	//在这种情况下，_sha1.Sum(nil) 会将 SHA1 哈希计算器的当前计算状态（即之前的输入数据）转换为最终的散列值。
	//这个方法通常用于计算文件的 SHA1 哈希值，因为在这种情况下，您通常不会在计算过程中添加任何额外的数据。
	//相反，您只是需要计算文件内容的 SHA1 哈希值。
	return hex.EncodeToString(_sha1.Sum(nil))
}

// MD5 用于计算数据的 MD5 哈希值
func MD5(data []byte) string {
	_md5 := md5.New()
	_md5.Write(data)
	return hex.EncodeToString(_md5.Sum([]byte("")))
}

// FileMD5 用于计算文件的 MD5 哈希值
func FileMD5(file *os.File) string {
	_md5 := md5.New()
	io.Copy(_md5, file)
	return hex.EncodeToString(_md5.Sum([]byte(nil)))
}

// PathExists 检查给定的路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetFileSize 用于获取文件的大小
func GetFileSize(filename string) int64 {
	var result int64
	//用于递归地遍历文件系统中的目录树，并对每个文件和目录执行一个函数。
	//这个函数可以用于执行多种操作，如计算文件大小、检查文件属性、执行文件操作等。
	filepath.Walk(filename, func(path string, info fs.FileInfo, err error) error {
		result = info.Size()
		return nil
	})
	return result
}
