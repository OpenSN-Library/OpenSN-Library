package handler

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/share/dir"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetFileList(ctx *gin.Context) {
	var req ginmodel.OperateFileReq
	err := ctx.Bind(&req)
	if err != nil {
		errMsg := fmt.Sprintf("Parse File Operation Request Error: %s", err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	listPath := path.Join(dir.UserDataDir, req.RelativePath)
	var fileList []*ginmodel.FileNode
	files, err := os.ReadDir(listPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Dir %s Error: %s", listPath, err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	for _, file := range files {
		fileNode := &ginmodel.FileNode{
			Name:      file.Name(),
			Privilege: int(file.Type().Perm()),
			IsDir:     file.IsDir(),
		}
		fileInfo, err := file.Info()
		if err == nil {
			fileNode.LastModifyTime = fileInfo.ModTime()
			fileNode.Size = fileInfo.Size()
		}
		fileList = append(fileList, fileNode)
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    fileList,
	}
	ctx.JSON(http.StatusOK, resp)
}

func UploadFile(ctx *gin.Context) {
	// 单文件
	relativePath := ctx.GetString("path")
	file, err := ctx.FormFile("file")
	if err != nil {
		errMsg := fmt.Sprintf("Get Upload File Error :%s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	dst := path.Join(dir.UserDataDir, relativePath)
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		errMsg := fmt.Sprintf("Save Upload File %s Error :%s", file.Filename, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    nil,
	}
	ctx.JSON(http.StatusOK, resp)
}

func DeleteFile(ctx *gin.Context) {
	var req ginmodel.OperateFileReq
	err := ctx.Bind(&req)
	if err != nil {
		errMsg := fmt.Sprintf("Parse File Operation Request Error: %s", err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	removePath := path.Join(dir.UserDataDir, req.RelativePath)
	err = os.RemoveAll(removePath)
	if err != nil {
		errMsg := fmt.Sprintf("Remove %s Error: %s", removePath, err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    nil,
	}
	ctx.JSON(http.StatusOK, resp)
}

func DownloadFile(ctx *gin.Context) {
	var req ginmodel.OperateFileReq
	err := ctx.Bind(&req)
	if err != nil {
		errMsg := fmt.Sprintf("Parse File Operation Request Error: %s", err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	downloadPath := path.Join(dir.UserDataDir, req.RelativePath)

	fileStat, err := os.Stat(downloadPath)

	if err != nil {
		errMsg := fmt.Sprintf("Locate Download File %s Error: %s", req.RelativePath, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	if fileStat.IsDir() {
		errMsg := "Download File Error: cannot download a dir"
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx.File(downloadPath)
}
