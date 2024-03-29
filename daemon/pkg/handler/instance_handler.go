package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AddInstanceHandler(ctx *gin.Context) {

}

func DelInstanceHandler(ctx *gin.Context) {
	var req ginmodel.SingleInstanceRequest
	if err := ctx.Bind(&req); err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Request Data: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}

	instance, err := synchronizer.GetInstanceInfo(req.NodeIndex, req.InstanceID)

	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Instance Info Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	for _, connection := range instance.Connections {
		if connection.EndNodeIndex == instance.NodeIndex {
			err = synchronizer.RemoveLink(instance.NodeIndex, connection.LinkID)
			if err != nil {
				errMsg := fmt.Sprintf("Delete Link %s Error: %s", connection.LinkID, err.Error())
				logrus.Error(errMsg)
			}
		} else {
			err = synchronizer.RemoveLink(connection.EndNodeIndex, connection.LinkID)
			if err != nil {
				errMsg := fmt.Sprintf("Delete Link %s Error: %s", connection.LinkID, err.Error())
				logrus.Error(errMsg)
			}
			err = synchronizer.RemoveLink(instance.NodeIndex, connection.LinkID)
			if err != nil {
				errMsg := fmt.Sprintf("Delete Link %s Error: %s", connection.LinkID, err.Error())
				logrus.Error(errMsg)
			}
		}
	}
	err = synchronizer.RemoveInstance(req.NodeIndex, req.InstanceID)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Remove Instance Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func GetInstanceListHandler(ctx *gin.Context) {
	var req ginmodel.GetInstanceListRequest
	var instanceList []*model.Instance
	var respData []ginmodel.InstanceAbstract

	if err := ctx.Bind(&req); err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Request Data: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	nodes, err := synchronizer.GetNodeList()
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get NodeList Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	for _, node := range nodes {
		nodeInstanceList, err := synchronizer.GetInstanceList(node.NodeIndex)
		if err != nil {
			errMsg := fmt.Sprintf("Get Instance List of Node %d Error: %s", node.NodeIndex, err.Error())
			logrus.Error(errMsg)
			continue
		}
		for _, instance := range nodeInstanceList {
			if req.KeyWord == "" || strings.Contains(instance.Name, req.KeyWord) {
				instanceList = append(instanceList, instance)
			}
		}
	}

	startIndex := 0
	endIndex := len(instanceList)

	if req.PageIndex >= 0 && req.PageSize > 0 {
		startIndex = req.PageSize * req.PageIndex
		endIndex = startIndex + req.PageSize
	}

	for i := startIndex; i < endIndex && i < len(instanceList); i++ {
		instance := instanceList[i]
		respData = append(respData, ginmodel.InstanceAbstract{
			InstanceID: instance.InstanceID,
			Name:       instance.Name,
			Type:       instance.Type,
			Start:      instance.Start,
			Extra:      instance.Extra,
		})
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)

}

func GetInstanceInfoHandler(ctx *gin.Context) {
	var respData ginmodel.InstanceInfo

	instanceID := ctx.Param("instance_id")
	nodeIndexStr := ctx.Param("node_index")
	nodeIndex, err := strconv.Atoi(nodeIndexStr)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Node Index: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	instance, err := synchronizer.GetInstanceInfo(nodeIndex, instanceID)

	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Instance Info Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	respData = ginmodel.InstanceInfo{
		InstanceID:    instance.InstanceID,
		Name:          instance.Name,
		Type:          instance.Type,
		Start:         instance.Start,
		Extra:         instance.Extra,
		Connections:   instance.Connections,
		Image:         instance.Image,
		ResourceLimit: instance.Resource,
		NodeIndex:     instance.NodeIndex,
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func StartInstanceHander(ctx *gin.Context) {

}

func StopInstanceHandler(ctx *gin.Context) {

}
