package httpserv

import (
	"VideoServ/glb"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"github.com/jsuserapp/ju"
	"github.com/kataras/iris/v12/mvc"
	"github.com/syndtr/goleveldb/leveldb"
	"math"
	"os"
	"path/filepath"
)

type ApiController struct {
	Controller
}

func (ac *ApiController) PostSourceList() mvc.Result {
	return ac.JsonRequest(nil, func(js ju.JsonObject) string {
		conf := glb.GetConf()
		sourceList := make([]struct {
			Hash string `json:"hash"`
			Path string `json:"path"`
		}, len(conf.VideoServer.SourceList))
		for i := range sourceList {
			source := conf.VideoServer.SourceList[i]
			sourceList[i].Path = glb.PathToWindows(source.Path)
			sourceList[i].Hash = source.Hash
		}
		js.SetValue("source_list", sourceList)
		return ""
	})
}
func (ac *ApiController) PostSourceListSave() mvc.Result {
	obj := struct {
		SourceList []struct {
			Path string `json:"path"`
			Hash string `json:"hash"`
		} `json:"source_list"`
	}{}
	return ac.JsonRequest(&obj, func(js ju.JsonObject) string {
		conf := glb.GetConf()
		for i, source := range obj.SourceList {
			obj.SourceList[i].Path = glb.PathToLinux(source.Path)
		}
		conf.VideoServer.SourceList = obj.SourceList
		if !conf.Save() {
			return "保存失败"
		}
		return ""
	})
}

type VideoInfo struct {
	Hash  string `json:"hash"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

func (ac *ApiController) PostVideoList() mvc.Result {
	obj := struct {
		Hash string `json:"hash"`
		Path string `json:"path"`
	}{}
	return ac.JsonRequest(&obj, func(js ju.JsonObject) string {
		conf := glb.GetConf()
		vis := make([]*VideoInfo, 0)
		if obj.Hash == "" {
			for _, source := range conf.VideoServer.SourceList {
				info, err := os.Stat(source.Path)
				if err != nil {
					continue
				}
				vis = addVideoItem(source.Hash, source.Path, "", info, vis)
			}
		} else {
			obj.Path = glb.PathToLinux(obj.Path)
			root := ""
			for _, source := range conf.VideoServer.SourceList {
				if source.Hash == obj.Hash {
					root = source.Path
					break
				}
			}
			if root == "" {
				return "无效源文件夹"
			}
			rootpath := filepath.Join(root, obj.Path)
			items, err := os.ReadDir(rootpath)
			if ju.CheckFailure(err) {
				return err.Error()
			}
			for _, item := range items {
				path := filepath.Join(obj.Path, item.Name())
				info, err := item.Info()
				if err != nil {
					continue
				}
				vis = addVideoItem(obj.Hash, root, path, info, vis)
			}
		}
		js.SetValue("video_list", vis)
		return ""
	})
}
func addVideoItem(hash, root, path string, info os.FileInfo, vis []*VideoInfo) []*VideoInfo {
	vi := &VideoInfo{
		Hash:  hash,
		Path:  glb.PathToWindows(path),
		Name:  info.Name(),
		IsDir: info.IsDir(),
		Size:  info.Size(),
	}
	if info.IsDir() {
		vis = append(vis, vi)
		return vis
	}
	if isVideoExt(info.Name()) {
		vis = append(vis, vi)
	}
	return vis
}
func (ac *ApiController) PostVideoPosition() mvc.Result {
	obj := struct {
		Hash string `json:"hash"`
		Path string `json:"path"`
	}{}
	return ac.JsonRequest(&obj, func(js ju.JsonObject) string {
		if obj.Hash == "" || obj.Path == "" {
			return "缺少参数"
		}
		obj.Path = glb.PathToLinux(obj.Path)
		conf := glb.GetConf()
		root := ""
		for _, source := range conf.VideoServer.SourceList {
			if source.Hash == obj.Hash {
				root = source.Path
				break
			}
		}
		if root == "" {
			return "无效源文件夹"
		}
		fullPath := filepath.Join(root, obj.Path)
		if glb.PathNotSafe(fullPath, js) {
			return "不合法的路径"
		}
		db := glb.GetDb()
		if db == nil {
			return "数据库未打开"
		}
		key := sha1.Sum([]byte(fullPath))
		val, err := db.Get(key[:], nil)
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				js.SetValue("position", 0)
				return ""
			}
			return err.Error()
		}
		if len(val) != 8 {
			return "错误的位置数据"
		}
		js.SetValue("position", math.Float64frombits(binary.BigEndian.Uint64(val)))
		return ""
	})
}
func (ac *ApiController) PostVideoPositionSave() mvc.Result {
	obj := struct {
		Hash     string  `json:"hash"`
		Path     string  `json:"path"`
		Position float64 `json:"position"`
	}{}
	return ac.JsonRequest(&obj, func(js ju.JsonObject) string {
		if obj.Hash == "" || obj.Path == "" {
			return "缺少参数"
		}
		conf := glb.GetConf()
		root := ""
		for _, source := range conf.VideoServer.SourceList {
			if source.Hash == obj.Hash {
				root = source.Path
				break
			}
		}
		if root == "" {
			return "无效源文件夹"
		}
		obj.Path = glb.PathToLinux(obj.Path)
		fullPath := filepath.Join(root, obj.Path)
		if glb.PathNotSafe(fullPath, js) {
			return "不合法的路径"
		}
		db := glb.GetDb()
		if db == nil {
			return "数据库未打开"
		}
		key := sha1.Sum([]byte(fullPath))

		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(obj.Position))
		err := db.Put(key[:], buf[:], nil)
		if ju.CheckFailure(err) {
			return err.Error()
		}
		return ""
	})
}
func (ac *ApiController) PostPlayPosition() mvc.Result {
	return ac.JsonRequest(nil, func(js ju.JsonObject) string {
		conf := glb.GetConf()
		js.SetValue("play_position", conf.PlayPosition)
		return ""
	})
}
func (ac *ApiController) PostPlayPositionSave() mvc.Result {
	obj := struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	}{}
	return ac.JsonRequest(&obj, func(js ju.JsonObject) string {
		conf := glb.GetConf()
		conf.PlayPosition.Start = obj.Start
		conf.PlayPosition.End = obj.End
		if !conf.Save() {
			return "保存失败"
		}

		return ""
	})
}
