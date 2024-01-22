package ginmodel

import "time"

type OperateFileReq struct {
	RelativePath string `json:"relative_path"`
}

type FileNode struct {
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	Privilege      int       `json:"privilege"`
	LastModifyTime time.Time `json:"last_modify_time"`
	IsDir          bool      `json:"is_dir"`
}
