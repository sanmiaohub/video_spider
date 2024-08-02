package platform

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"os"
	"strings"
)

// Record 临时解析数据
type Record struct {
	Link         string      `json:"link"`
	Type         int         `json:"type"`
	Title        string      `json:"title"`
	Cover        string      `json:"cover"`
	Video        string      `json:"video"`
	ResourcePath interface{} `json:"resource_path"`
}

type Platform interface {
	ParseOut() (record Record, err error)
}

func getDecompressedReader(res *http.Response) (io.ReadCloser, error) {
	switch strings.ToLower(res.Header.Get("Content-Encoding")) {
	case "gzip":
		return gzip.NewReader(res.Body)
	case "deflate":
		return zlib.NewReader(res.Body)
	default:
		return res.Body, nil
	}
}

func debugFile(str []byte, filename string) {
	// 打开或创建一个文件用于写入
	file, err := os.Create(filename + ".json")
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
		}
	}(file)

	// 创建一个缓冲写入器
	writer := bufio.NewWriter(file)

	// 将文本写入缓冲区
	_, err = writer.WriteString(string(str))
	if err != nil {
		return
	}

	// 刷新缓冲区，确保所有数据都被写入文件
	err = writer.Flush()
	if err != nil {
		return
	}
}
