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

func AddLinkHandler(ctx *gin.Context) {

}

func DelLinkHandler(ctx *gin.Context) {

}

func GetLinkListHandler(ctx *gin.Context) {
	var req ginmodel.GetLinkListRequest
	var linkList []model.Link
	var respData []ginmodel.LinkAbstract
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
			Message: fmt.Sprintf("Get Node List Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	for _, node := range nodes {

		nodeLinkList, err := synchronizer.GetLinkList(node.NodeIndex)
		if err != nil {
			errMsg := fmt.Sprintf("Get Link List of Node %d Error: %s", node.NodeIndex, err.Error())
			logrus.Error(errMsg)
			continue
		}
		for _, link := range nodeLinkList {
			if req.KeyWord == "" || strings.Contains(link.GetLinkID(), req.KeyWord) {
				linkList = append(nodeLinkList, link)
			}
		}
	}
	startIndex := 0
	endIndex := len(linkList)

	if req.PageIndex >= 0 && req.PageSize > 0 {
		startIndex = req.PageSize * req.PageIndex
		endIndex = startIndex + req.PageSize
	}

	for i := startIndex; i < endIndex && i < len(linkList); i++ {
		link := linkList[i]
		linkAbstract := ginmodel.LinkAbstract{
			LinkID: link.GetLinkID(),
			Type:   link.GetLinkType(),
			Enable: link.GetLinkBasePtr().Enable,
			ConnectIntance: [2]string{
				link.GetLinkBasePtr().EndInfos[0].InstanceID,
				link.GetLinkBasePtr().EndInfos[1].InstanceID,
			},
		}
		respData = append(respData, linkAbstract)
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)

}

func GetLinkInfoHandler(ctx *gin.Context) {
	var respData ginmodel.LinkInfo
	linkID := ctx.Param("link_id")
	nodeIndexStrs := strings.Split(ctx.Param("node_index"), "+")
	nodeIndex, err := strconv.Atoi(nodeIndexStrs[0])
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Node Index: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	link, err := synchronizer.GetLinkInfo(nodeIndex, linkID)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Link Info of %s Error: %s", linkID, err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	respData.LinkID = link.GetLinkID()
	respData.Type = link.GetLinkType()
	respData.Enable = link.GetLinkBasePtr().Enable
	for i := 0; i < len(link.AddressInfos); i++ {
		respData.ConnectIntance[i] = link.EndInfos[i].InstanceID
		respData.AddressInfos[i] = link.AddressInfos[i]
	}
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func GetLinkParameterListHandler(ctx *gin.Context) {
	respData := map[string]map[string]int64{}
	nodes, err := synchronizer.GetNodeList()
	for _, node := range nodes {
		parameters, err := synchronizer.GetLinkListParameters(node.NodeIndex)
		if err != nil {
			errMsg := fmt.Sprintf("Get Link Parameter List of Node %d Error: %s", node.NodeIndex, err.Error())
			logrus.Error(errMsg)
			continue
		}
		for k, v := range parameters {
			respData[k] = v
		}
	}
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Link Parameter List Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func GetLinkParameterHandler(ctx *gin.Context) {
	var req ginmodel.SingleLinkRequest
	var respData map[string]int64
	if err := ctx.Bind(&req); err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Request Data: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	parameter, err := synchronizer.GetLinkParameter(req.NodeIndex, req.LinkID)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Link Parameter of %s Error: %s", req.LinkID, err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	respData = parameter
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func UpdateLinkParameterHandler(ctx *gin.Context) {

}
