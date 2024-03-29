package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartInstanceWebshellHandler(ctx *gin.Context) {
	var req ginmodel.InstanceWebshellRequest
	if err := ctx.Bind(&req); err != nil {
		errMsg := fmt.Sprintf("Bind Webshell Request Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}

	webShellID := fmt.Sprintf("instance_%s", req.InstanceID)

	info, err := synchronizer.GetWebshellInfo(req.NodeIndex, webShellID)

	if err == nil {
		ctx.JSON(http.StatusOK, ginmodel.JsonResp{
			Code:    0,
			Message: "Success",
			Data:    info,
		})
		return
	}

	instanceInfo, err := synchronizer.GetInstanceInfo(req.NodeIndex, req.InstanceID)

	if err != nil {
		errMsg := fmt.Sprintf("Get Instance Info Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}

	requestPayload := model.WebShellAllocRequest{
		WebShellID:   webShellID,
		Command:      "docker",
		Args:         []string{"exec", "-it", fmt.Sprintf("%s_%s", instanceInfo.Type, req.InstanceID), "/bin/bash"},
		Writeable:    true,
		ExpireMinute: req.ExpireMinute,
	}

	err = synchronizer.UpdateWebShellRequest(req.NodeIndex, &requestPayload)
	if err != nil {
		errMsg := fmt.Sprintf("Update Webshell Request Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}
	info, err = synchronizer.WaitWebshellInfo(req.NodeIndex, requestPayload.WebShellID, time.Minute)
	if err != nil {
		errMsg := fmt.Sprintf("Wait Webshell Info Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    info,
	})
}

func StartLinkWebshellHandler(ctx *gin.Context) {
	var req ginmodel.LinkWebshellRequest
	if err := ctx.Bind(&req); err != nil {
		errMsg := fmt.Sprintf("Bind Webshell Request Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}

	webShellID := fmt.Sprintf("link_%s", req.LinkID)

	info, err := synchronizer.GetWebshellInfo(req.NodeIndex, webShellID)
	if err == nil {
		ctx.JSON(http.StatusOK, ginmodel.JsonResp{
			Code:    0,
			Message: "Success",
			Data:    info,
		})
		return
	}

	_, err = synchronizer.GetLinkInfo(req.NodeIndex, req.LinkID)

	if err != nil {
		errMsg := fmt.Sprintf("Get Link Info Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}

	requestPayload := model.WebShellAllocRequest{
		WebShellID:   webShellID,
		Command:      "tcpdump",
		Args:         []string{"-e", "-i", req.LinkID},
		Writeable:    true,
		ExpireMinute: req.ExpireMinute,
	}

	err = synchronizer.UpdateWebShellRequest(req.NodeIndex, &requestPayload)
	if err != nil {
		errMsg := fmt.Sprintf("Update Webshell Request Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}
	info, err = synchronizer.WaitWebshellInfo(req.NodeIndex, requestPayload.WebShellID, time.Minute)
	if err != nil {
		errMsg := fmt.Sprintf("Wait Webshell Info Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
		return
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    info,
	})
}
