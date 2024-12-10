package httpserv

import (
	"VideoServ/glb"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"io"
	"ju"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type VideoController struct {
	Controller
}

func (vc *VideoController) GetBy(hash, path string) mvc.Result {
	conf := glb.GetConf()
	root := ""
	for _, source := range conf.VideoServer.SourceList {
		if source.Hash == hash {
			root = source.Path
			break
		}
	}
	// 如果根路径或子路径为空，返回错误
	if root == "" || path == "" {
		return vc.Err(http.StatusBadRequest, "Invalid hash or path")
	}
	fullPath := filepath.Join(root, path)
	fullPath = glb.PathToLinux(fullPath)
	// 检查文件是否存在
	file, err := os.Open(fullPath)
	if err != nil {
		return vc.Err(iris.StatusNotFound, "文件打开出错")
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return vc.Err(iris.StatusInternalServerError, "获取文件信息出错")
	}

	fileSize := fileInfo.Size()
	startByte := int64(0)
	endByte := fileSize - 1

	// 解析Range header, 如果存在的话.
	rangeHeader := vc.Ctx.GetHeader("Range")
	if rangeHeader != "" {
		// 处理Range请求
		startByte, endByte, err = parseRange(rangeHeader, fileSize)
	}

	mime, err := mimetype.DetectFile(fullPath)

	ju.CheckError(err)
	// 设置响应头
	vc.Ctx.Header("Content-Type", mime.String())
	vc.Ctx.Header("Accept-Ranges", "bytes")

	if startByte > 0 || endByte < fileSize-1 {
		vc.Ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startByte, endByte, fileSize))
		vc.Ctx.StatusCode(iris.StatusPartialContent)
	} else {
		vc.Ctx.StatusCode(iris.StatusOK)
	}

	// 设置Content-Length
	vc.Ctx.Header("Content-Length", strconv.FormatInt(endByte-startByte+1, 10))

	// 流式传输文件内容
	_, err = io.CopyN(vc.Ctx.ResponseWriter(), io.NewSectionReader(file, startByte, endByte-startByte+1), endByte-startByte+1)
	if err != nil && err != io.EOF {
		return vc.Err(iris.StatusInternalServerError, "IO 错误")
	}

	return nil
}
func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	headFlag := "bytes="
	// 如果没有提供Range头，则返回整个文件
	if rangeHeader == "" {
		return 0, fileSize - 1, nil
	}
	if strings.Index(rangeHeader, headFlag) != 0 {
		return 0, fileSize - 1, fmt.Errorf("%s", rangeHeader)
	}
	rangeHeader = rangeHeader[len(headFlag):]
	parts := strings.Split(rangeHeader, "-")
	if len(parts) != 2 {
		return 0, fileSize - 1, fmt.Errorf("%s", rangeHeader)
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= fileSize {
		return 0, fileSize - 1, fmt.Errorf(rangeHeader)
	}

	end := fileSize - 1
	if parts[1] != "" {
		end64, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end64 < start || end64 >= fileSize {
			return 0, fileSize - 1, fmt.Errorf("%s", rangeHeader)
		}
		end = end64
	}

	return start, end, nil
}
