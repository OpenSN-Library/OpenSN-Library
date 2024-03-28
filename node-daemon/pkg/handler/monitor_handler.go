package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetNodeResourceDataHandler(ctx *gin.Context) {
	var nodeIndexes []int
	nodeIndexStr := ctx.Param("node_index")
	if nodeIndexStr == "all" || nodeIndexStr == "" {
		nodes, err := synchronizer.GetNodeList()
		if err != nil {
			errMsg := "Get Node List Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
		}
		for _, node := range nodes {
			nodeIndexes = append(nodeIndexes, node.NodeIndex)
		}
	} else {
		nodeIndex, err := strconv.Atoi(nodeIndexStr)
		if err != nil {
			errMsg := "Invalid Node Index: " + err.Error()
			ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			nodeIndexes = append(nodeIndexes, nodeIndex)
		}
	}
	hostResources, err := synchronizer.GetLastNodeResourceDatas(nodeIndexes)
	if err != nil {
		errMsg := "Get Last Node Resource Data Error: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		})
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    hostResources,
	})
}

func GetLinkResourceDataHander(ctx *gin.Context) {
	var respData map[string]*model.LinkResource
	var err error
	linkIDStr := ctx.Param("link_id")
	if linkIDStr == "" || linkIDStr == "all" {
		respData, err = synchronizer.GetLastAllLinkResourceDatas()
		if err != nil {
			errMsg := "Get Last All Link Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
	} else {
		linkResource, err := synchronizer.GetLastLinkResourceDatas([]string{linkIDStr})
		if err != nil {
			errMsg := "Get Last Link Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})

		}
		respData = map[string]*model.LinkResource{
			linkIDStr: &linkResource[0],
		}

	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	})
}

func GetInstanceResourceDataHandler(ctx *gin.Context) {
	var respData map[string]*model.InstanceResouce
	var err error
	instanceIDStr := ctx.Param("instance_id")
	if instanceIDStr == "" || instanceIDStr == "all" {
		respData, err = synchronizer.GetLastAllInstanceResourceDatas()
		if err != nil {
			errMsg := "Get Last All Instance Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
	} else {
		instanceResource, err := synchronizer.GetLastInstanceResourceDatas([]string{instanceIDStr})
		if err != nil {
			errMsg := "Get Last Instance Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})

		}
		respData = map[string]*model.InstanceResouce{
			instanceIDStr: &instanceResource[0],
		}
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	})
}

func GetPeriodNodeResourceDataHandler(ctx *gin.Context) {
	periodExpr := ctx.Query("period")
	nodeIndexStr := ctx.Param("node_index")
	hostResourceMap := make(map[int][]*model.HostResource)
	if nodeIndexStr == "" || nodeIndexStr == "all" {
		nodes, err := synchronizer.GetNodeList()
		if err != nil {
			errMsg := "Get Node List Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		var nodeIndexes []int
		for _, node := range nodes {
			nodeIndexes = append(nodeIndexes, node.NodeIndex)
		}
		hostResources, err := synchronizer.GetPeriodNodeResourceDatas(periodExpr, nodeIndexes)
		if err != nil {
			errMsg := "Get Period Node Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}

		for i, nodeIndex := range nodeIndexes {
			hostResourceMap[nodeIndex] = hostResources[i]
		}

	} else {
		nodeIndex, err := strconv.Atoi(nodeIndexStr)
		if err != nil {
			errMsg := "Invalid Node Index: " + err.Error()
			ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		hostResources, err := synchronizer.GetPeriodNodeResourceDatas(periodExpr, []int{nodeIndex})
		if err != nil {
			errMsg := "Get Period Node Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		hostResourceMap[nodeIndex] = hostResources[0]
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    hostResourceMap,
	})
}

func GetPeriodInstanceResourceDataHandler(ctx *gin.Context) {
	periodExpr := ctx.Query("period")
	instanceIDStr := ctx.Param("instance_id")
	instanceResourceMap := make(map[string][]*model.InstanceResouce)
	if instanceIDStr == "" || instanceIDStr == "all" {
		instanceResources, err := synchronizer.GetPeriodAllInstanceResourceDatas(periodExpr)
		if err != nil {
			errMsg := "Get Period All Instance Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		instanceResourceMap = instanceResources
	} else {
		instanceResources, err := synchronizer.GetPeriodInstanceResourceDatas(periodExpr, instanceIDStr)
		if err != nil {
			errMsg := "Get Period Instance Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		instanceResourceMap[instanceIDStr] = instanceResources
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    instanceResourceMap,
	})
}

func GetPeriodLinkResourceDataHander(ctx *gin.Context) {
	periodExpr := ctx.Query("period")
	linkIDStr := ctx.Param("link_id")
	linkResourceMap := make(map[string][]*model.LinkResource)
	if linkIDStr == "" || linkIDStr == "all" {
		linkResources, err := synchronizer.GetPeriodAllLinkResourceDatas(periodExpr)
		if err != nil {
			errMsg := "Get Period All Link Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		linkResourceMap = linkResources
	} else {
		linkResources, err := synchronizer.GetPeriodLinkResourceDatas(periodExpr, linkIDStr)
		if err != nil {
			errMsg := "Get Period Link Resource Data Error: " + err.Error()
			ctx.JSON(http.StatusInternalServerError, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			})
			return
		}
		linkResourceMap[linkIDStr] = linkResources
	}
	ctx.JSON(http.StatusOK, ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    linkResourceMap,
	})
}