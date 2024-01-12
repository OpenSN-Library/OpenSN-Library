package handler

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/share/data"
	"NodeDaemon/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetNodeListHandler(ctx *gin.Context) {
	var nodeList []ginmodel.NodeAbstract

	for _, v := range data.NodeMap {
		nodeList = append(nodeList, ginmodel.NodeAbstract{
			NodeID:       v.NodeID,
			FreeInstance: v.FreeInstance,
			IsMasterNode: v.IsMasterNode,
			L3AddrV4:     utils.FormatIPv4(v.L3AddrV4),
			L3AddrV6:     utils.FormatIPv6(v.L3AddrV6),
			L2Addr:       utils.FormatMacAddr(v.L2Addr),
		})
	}

	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    nodeList,
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

	v, ok := data.NodeMap[nodeIndex]

	if !ok {
		errMsg := fmt.Sprintf("Node %d Not Found.", nodeIndex)
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusNotFound, jsonResp)
		return
	}

	var obj = ginmodel.NodeDetail{
		NodeAbstract: ginmodel.NodeAbstract{
			NodeID:       v.NodeID,
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
