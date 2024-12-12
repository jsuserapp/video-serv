package httpserv

import (
	"bytes"
	"github.com/jsuserapp/ju"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".flv":  true,
	".wmv":  true,
	".webm": true,
	".dat":  true,
	".m2ts": true,
	".ts":   true,
	".rmvb": true,
	".rm":   true,
}

// 判断文件是否是视频文件
func isVideoExt(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return videoExtensions[ext]
}
func isVideoFileByMimeType(filePath string) bool {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	// 读取文件的前 512 个字节（用于 MIME 类型检测）
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	// 检测 MIME 类型
	mimeType := http.DetectContentType(buffer)
	ju.LogCyan(mimeType)
	return 0 == strings.Index(mimeType, "video/")
}

var videoMagicNumbers = map[string][]byte{
	"MPEG-PS": []byte{0x00, 0x00, 0x01, 0xBA}, // VCD 视频或 .dat 文件
	"MPEG-TS": []byte{0x47},                   // 蓝光 .m2ts 文件或 TS 文件
	"MP4":     []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70},
	"MKV":     []byte{0x1A, 0x45, 0xDF, 0xA3},
	"AVI":     []byte{0x52, 0x49, 0x46, 0x46},
	"FLV":     []byte{0x46, 0x4C, 0x56},
}

// 判断文件是否是视频文件
func isVideoFileByMagicNumber(filePath string) (bool, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, "", err
	}
	defer func() {
		_ = file.Close()
	}()

	// 读取文件头
	buffer := make([]byte, 16) // 读取前 16 个字节
	_, err = file.Read(buffer)
	if err != nil {
		return false, "", err
	}

	// 匹配 Magic Number
	for format, magic := range videoMagicNumbers {
		if bytes.HasPrefix(buffer, magic) {
			return true, format, nil
		}
	}

	return false, "", nil
}
