package glb

import (
	"github.com/jsuserapp/ju"
	"strings"
)

func PathNotSafe(path string, js ju.JsonObject) bool {
	if strings.Index(path, "../") != -1 || strings.Index(path, "/..") != -1 {
		js.SetValue("error", "文件路径错误")
		return true
	}
	return false
}
func PathToLinux(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}
func PathToWindows(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}
