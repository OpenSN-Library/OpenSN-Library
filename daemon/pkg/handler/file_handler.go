package handler

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/share/dir"
	"NodeDaemon/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func PreviewFileHandler(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	previewPath := path.Join(dir.UserDataDir, relativePath)
	fileType, err := utils.CheckPathType(previewPath)
	var previewText string
	if err != nil {
		errMsg := fmt.Sprintf("Check File Type Error: %s", err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	switch fileType {
	case utils.DirType:
		previewText = "不支持预览文件夹，请下载查看"
	case utils.BinaryType:
		previewText = "不支持预览二进制文件，请下载查看"
	case utils.TextType:
		f, err := os.Open(previewPath)
		if err != nil {
			errMsg := fmt.Sprintf("Open File %s Error: %s", previewPath, err.Error())
			logrus.Errorf(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
		defer f.Close()
		buf, err := io.ReadAll(f)
		if err != nil {
			errMsg := fmt.Sprintf("Read File %s Error: %s", previewPath, err.Error())
			logrus.Errorf(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
		previewText = string(buf)
	default:
		previewText = "未知文件类型,请下载查看"
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    previewText,
	}
	ctx.JSON(http.StatusOK, resp)
}

func GetFileListHandler(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	listPath := path.Join(dir.UserDataDir, relativePath)
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

func UploadFileHandler(ctx *gin.Context) {
	// 单文件
	relativePath := ctx.Query("path")
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
	dst := path.Join(dir.UserDataDir, relativePath, file.Filename)
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

func DeleteFileHandler(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	removePath := path.Join(dir.UserDataDir, relativePath)
	err := os.RemoveAll(removePath)
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

func DownloadFileHandler(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	downloadPath := path.Join(dir.UserDataDir, relativePath)

	fileStat, err := os.Stat(downloadPath)

	if err != nil {
		errMsg := fmt.Sprintf("Locate Download File %s Error: %s", relativePath, err.Error())
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
