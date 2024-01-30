package handler

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetNodeListHandler(ctx *gin.Context) {

	nodeList, err := synchronizer.GetNodeList()

	if err != nil {
		errMsg := fmt.Sprintf("Get Node List From Etcd Error:%s", err.Error())
		logrus.Error(errMsg)
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	var nodeAbstractList []ginmodel.NodeAbstract

	for _, nodeInfo := range nodeList {
		nodeAbstractList = append(nodeAbstractList, ginmodel.NodeAbstract{
			NodeIndex:    nodeInfo.NodeIndex,
			FreeInstance: nodeInfo.FreeInstance,
			IsMasterNode: nodeInfo.IsMasterNode,
			L3AddrV4:     utils.FormatIPv4(nodeInfo.L3AddrV4),
			L3AddrV6:     utils.FormatIPv6(nodeInfo.L3AddrV6),
			L2Addr:       utils.FormatMacAddr(nodeInfo.L2Addr),
		})
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    nodeAbstractList,
	}

	ctx.JSON(http.StatusOK, jsonResp)
}

func GetNodeInfoHandler(ctx *gin.Context) {
	nodeIndexStr := ctx.Param("node_index")

	nodeIndex, err := strconv.Atoi(nodeIndexStr)

	if err != nil {
		errMsg := fmt.Sprintf("Parse node index %s Error:%s", nodeIndexStr, err.Error())
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}

	v, err := synchronizer.GetNode(nodeIndex)

	if err != nil {
		errMsg := fmt.Sprintf("Get Node Info of %d Error:%s", nodeIndex, err.Error())
		logrus.Error(errMsg)
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}

	var obj = ginmodel.NodeDetail{
		NodeAbstract: ginmodel.NodeAbstract{
			NodeIndex:    v.NodeIndex,
			FreeInstance: v.FreeInstance,
			IsMasterNode: v.IsMasterNode,
			L3AddrV4:     utils.FormatIPv4(v.L3AddrV4),
			L3AddrV6:     utils.FormatIPv6(v.L3AddrV6),
			L2Addr:       utils.FormatMacAddr(v.L2Addr),
		},
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    obj,
	}

	ctx.JSON(http.StatusOK, jsonResp)

}
