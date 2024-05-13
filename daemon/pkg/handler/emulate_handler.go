package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/arranger"
	"NodeDaemon/pkg/link"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func GetEmulationConfigHandler(ctx *gin.Context) {
	var data ginmodel.EmulationDetail
	emulationConfig, err := synchronizer.GetEmulationInfo()
	if err != nil {
		errMsg := fmt.Sprintf("Get Emulation Config Error: get emulation info error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	data.Running = emulationConfig.Running
	data.InstanceTypeConfig = emulationConfig.TypeConfig
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    emulationConfig,
	}
	ctx.JSON(http.StatusOK, resp)
}

func ConfigEmulationHandler(ctx *gin.Context) {
	var req ginmodel.ConfigEmulationReq
	err := ctx.Bind(&req)
	if err != nil {
		errMsg := fmt.Sprintf("Config Emulation Error: parse request error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	err = synchronizer.UpdateEmulationInfo(func(ei *model.EmulationInfo) error {
		ei.TypeConfig = make(map[string]model.InstanceTypeConfig)
		for k, v := range req {
			nanoCPU, err := utils.ParseDecNumber(v.ResourceLimit.NanoCPU)
			if err != nil {
				return fmt.Errorf("parse nanocpu info error: %s", err.Error())
			}
			memByte, err := utils.ParseBinNumber(v.ResourceLimit.MemoryByte)
			if err != nil {
				return fmt.Errorf("parse memorybyte info error: %s", err.Error())
			}
			ei.TypeConfig[k] = model.InstanceTypeConfig{
				Image: v.Image,
				Envs:  v.Envs,
				ResourceLimit: model.ResourceLimit{
					NanoCPU:    nanoCPU,
					MemoryByte: memByte,
				},
			}
		}
		return nil
	})

	if err != nil {
		errMsg := fmt.Sprintf("Update Emulation Config Error: update emulation info error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, resp)

}

func StartEmulationHandler(ctx *gin.Context) {
	nodeList, err := synchronizer.GetNodeList()
	if err != nil {
		errMsg := fmt.Sprintf("Start Emulation Error: get node list error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	for _, nodeInfo := range nodeList {
		go func(nodeInfo *model.Node) {
			instanceList, err := synchronizer.GetInstanceList(nodeInfo.NodeIndex)

			if err != nil {
				errMsg := fmt.Sprintf("Start Emulation Error: get instance list of %d error: %s", nodeInfo.NodeIndex, err.Error())
				logrus.Error(errMsg)
				// resp := ginmodel.JsonResp{
				// 	Code:    -1,
				// 	Message: errMsg,
				// }
				// ctx.JSON(http.StatusInternalServerError, resp)
				// return
			}
			for _, instanceInfo := range instanceList {
				err := synchronizer.UpdateInstanceInfo(
					nodeInfo.NodeIndex,
					instanceInfo.InstanceID,
					func(i *model.Instance) error {
						i.Start = true
						return nil
					},
				)
				if err != nil {
					errMsg := fmt.Sprintf("Start Emulation Error: update instance %s state error: %s", instanceInfo.InstanceID, err.Error())
					logrus.Error(errMsg)
					// resp := ginmodel.JsonResp{
					// 	Code:    -1,
					// 	Message: errMsg,
					// }
					// ctx.JSON(http.StatusInternalServerError, resp)
					// return
				}
			}
		}(nodeInfo)
		go func(nodeInfo *model.Node) {
			linkList, err := synchronizer.GetLinkList(nodeInfo.NodeIndex)

			if err != nil {
				errMsg := fmt.Sprintf("Start Emulation Error: get link list of %d error: %s", nodeInfo.NodeIndex, err.Error())
				logrus.Error(errMsg)
				// resp := ginmodel.JsonResp{
				// 	Code:    -1,
				// 	Message: errMsg,
				// }
				// ctx.JSON(http.StatusInternalServerError, resp)
				// return
			}
			for _, linkInfo := range linkList {
				err := synchronizer.UpdateLinkInfo(
					nodeInfo.NodeIndex,
					linkInfo.GetLinkID(),
					func(i *model.LinkBase) error {
						i.Enable = true
						return nil
					},
				)
				if err != nil {
					errMsg := fmt.Sprintf("Start Emulation Error: update link %s state error: %s", linkInfo.GetLinkID(), err.Error())
					logrus.Error(errMsg)
					// resp := ginmodel.JsonResp{
					// 	Code:    -1,
					// 	Message: errMsg,
					// }
					// ctx.JSON(http.StatusInternalServerError, resp)
					// return
				}
			}
		}(nodeInfo)
	}

	err = synchronizer.UpdateEmulationInfo(func(ei *model.EmulationInfo) error {
		ei.Running = true
		return nil
	})

	if err != nil {
		errMsg := fmt.Sprintf("Start Emulation Error: update emulation info error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, resp)
}

func StopEmulationHandler(ctx *gin.Context) {
	nodeList, err := synchronizer.GetNodeList()
	if err != nil {
		errMsg := fmt.Sprintf("Stop Emulation Error: get node list error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	for _, nodeInfo := range nodeList {
		go func(nodeInfo *model.Node) {
			instanceList, err := synchronizer.GetInstanceList(nodeInfo.NodeIndex)
			if err != nil {
				errMsg := fmt.Sprintf("Stop Emulation Error: get instance list of %d error: %s", nodeInfo.NodeIndex, err.Error())
				logrus.Error(errMsg)
				// resp := ginmodel.JsonResp{
				// 	Code:    -1,
				// 	Message: errMsg,
				// }
				// ctx.JSON(http.StatusInternalServerError, resp)
				// return
			}
			for _, instanceInfo := range instanceList {
				synchronizer.UpdateInstanceInfo(
					nodeInfo.NodeIndex,
					instanceInfo.InstanceID,
					func(i *model.Instance) error {
						i.Start = false
						return nil
					},
				)
			}
		}(nodeInfo)

		go func(nodeInfo *model.Node) {
			linkList, err := synchronizer.GetLinkList(nodeInfo.NodeIndex)

			if err != nil {
				errMsg := fmt.Sprintf("Stop Emulation Error: get link list of %d error: %s", nodeInfo.NodeIndex, err.Error())
				logrus.Error(errMsg)
				// resp := ginmodel.JsonResp{
				// 	Code:    -1,
				// 	Message: errMsg,
				// }
				// ctx.JSON(http.StatusInternalServerError, resp)
				// return
			}
			for _, linkInfo := range linkList {
				err := synchronizer.UpdateLinkInfo(
					nodeInfo.NodeIndex,
					linkInfo.GetLinkID(),
					func(i *model.LinkBase) error {
						i.Enable = false
						return nil
					},
				)
				if err != nil {
					errMsg := fmt.Sprintf("Stop Emulation Error: update link %s state error: %s", linkInfo.GetLinkID(), err.Error())
					logrus.Error(errMsg)
					// resp := ginmodel.JsonResp{
					// 	Code:    -1,
					// 	Message: errMsg,
					// }
					// ctx.JSON(http.StatusInternalServerError, resp)
					// return
				}
			}
		}(nodeInfo)
	}

	err = synchronizer.UpdateEmulationInfo(func(ei *model.EmulationInfo) error {
		ei.Running = false
		return nil
	})

	if err != nil {
		errMsg := fmt.Sprintf("Stop Emulation Error: update emulation info error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, resp)
}

func AddTopologyHandler(ctx *gin.Context) {
	var req ginmodel.AddTopologyReq
	err := ctx.Bind(&req)
	if err != nil {
		errMsg := fmt.Sprintf("Add Topology Error: parse request error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	emulationConfig, err := synchronizer.GetEmulationInfo()
	if err != nil {
		errMsg := fmt.Sprintf("Add Topology Error: get emulation info error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	var instanceList []*model.Instance
	var linkList []*model.LinkBase
	for _, instance := range req.Instances {
		typeConfig, ok := emulationConfig.TypeConfig[instance.Type]
		if !ok {
			errMsg := fmt.Sprintf("Add Topology Error: type %s not in type set", instance.Type)
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusBadRequest, resp)
		}
		instanceID := uuid.NewString()[:8]
		instanceList = append(instanceList, &model.Instance{
			InstanceID:  instanceID,
			Name:        fmt.Sprintf("%s_%s", instance.Type, instanceID),
			Type:        instance.Type,
			Image:       typeConfig.Image,
			Extra:       instance.Extra,
			DeviceInfo:  instance.DeviceInfo,
			Resource:    typeConfig.ResourceLimit,
			Start:       emulationConfig.Running,
			Connections: make(map[string]model.ConnectionInfo),
		})
	}

	err = arranger.ArrangeInstances(instanceList)
	if err != nil {
		errMsg := fmt.Sprintf("Add Topology Error: arrange instance error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	for _, linkConfig := range req.Links {
		linkIndex := link.AllocLinkIndex()
		if linkConfig.EndIndexes[0] < 0 || linkConfig.EndIndexes[0] > len(instanceList) {
			continue
		}

		if linkConfig.EndIndexes[1] < 0 || linkConfig.EndIndexes[1] > len(instanceList) {
			continue
		}
		linkID := uuid.NewString()[:8]
		linkList = append(linkList, &model.LinkBase{
			LinkID: linkID,
			Enable: emulationConfig.Running,
			EndInfos: [2]model.EndInfoType{
				{
					InstanceID:   instanceList[linkConfig.EndIndexes[0]].InstanceID,
					InstanceType: instanceList[linkConfig.EndIndexes[0]].Type,
					EndNodeIndex: instanceList[linkConfig.EndIndexes[0]].NodeIndex,
				},
				{
					InstanceID:   instanceList[linkConfig.EndIndexes[1]].InstanceID,
					InstanceType: instanceList[linkConfig.EndIndexes[1]].Type,
					EndNodeIndex: instanceList[linkConfig.EndIndexes[1]].NodeIndex,
				},
			},
			LinkIndex:    linkIndex,
			Type:         linkConfig.Type,
			CrossMachine: instanceList[linkConfig.EndIndexes[0]].NodeIndex != instanceList[linkConfig.EndIndexes[1]].NodeIndex,
			NodeIndex:    instanceList[linkConfig.EndIndexes[0]].NodeIndex,
			Extra:        linkConfig.Extra,
			AddressInfos: linkConfig.AddressInfos,
		})
		if linkConfig.InitParameter == nil {
			linkList[len(linkList)-1].Parameter = make(map[string]int64)
		} else {
			linkList[len(linkList)-1].Parameter = linkConfig.InitParameter
		}
		if instanceList[linkConfig.EndIndexes[0]].NodeIndex != instanceList[linkConfig.EndIndexes[1]].NodeIndex {
			linkList = append(linkList, &model.LinkBase{
				LinkID: linkID,
				Enable: emulationConfig.Running,
				EndInfos: [2]model.EndInfoType{
					{
						InstanceID:   instanceList[linkConfig.EndIndexes[0]].InstanceID,
						InstanceType: instanceList[linkConfig.EndIndexes[0]].Type,
						EndNodeIndex: instanceList[linkConfig.EndIndexes[0]].NodeIndex,
					},
					{
						InstanceID:   instanceList[linkConfig.EndIndexes[1]].InstanceID,
						InstanceType: instanceList[linkConfig.EndIndexes[1]].Type,
						EndNodeIndex: instanceList[linkConfig.EndIndexes[1]].NodeIndex,
					},
				},
				LinkIndex:    linkIndex,
				Type:         linkConfig.Type,
				CrossMachine: instanceList[linkConfig.EndIndexes[0]].NodeIndex != instanceList[linkConfig.EndIndexes[1]].NodeIndex,
				Parameter:    linkConfig.InitParameter,
				NodeIndex:    instanceList[linkConfig.EndIndexes[1]].NodeIndex,
				Extra:        linkConfig.Extra,
				AddressInfos: linkConfig.AddressInfos,
			})
		}
		instanceList[linkConfig.EndIndexes[0]].Connections[linkID] = model.ConnectionInfo{
			LinkID:       linkID,
			InstanceID:   instanceList[linkConfig.EndIndexes[1]].InstanceID,
			EndNodeIndex: instanceList[linkConfig.EndIndexes[1]].NodeIndex,
			InstanceType: instanceList[linkConfig.EndIndexes[1]].Type,
		}
		instanceList[linkConfig.EndIndexes[1]].Connections[linkID] = model.ConnectionInfo{
			LinkID:       linkID,
			InstanceID:   instanceList[linkConfig.EndIndexes[0]].InstanceID,
			EndNodeIndex: instanceList[linkConfig.EndIndexes[0]].NodeIndex,
			InstanceType: instanceList[linkConfig.EndIndexes[0]].Type,
		}
	}

	wg := utils.ForEachWithThreadPool[*model.LinkBase](func(linkBase *model.LinkBase) {
		err := synchronizer.AddLinkInfo(linkBase.NodeIndex, linkBase)
		if err != nil {
			errMsg := fmt.Sprintf("Add Topology Error: add link %s to node %d error: %s",
				linkBase.LinkID,
				linkBase.NodeIndex,
				err.Error())
			logrus.Error(errMsg)
		}
	}, linkList, 512)
	wg.Wait()
	wg = utils.ForEachWithThreadPool[*model.Instance](func(instance *model.Instance) {
		err := synchronizer.AddInstanceInfo(instance.NodeIndex, instance)
		if err != nil {
			errMsg := fmt.Sprintf("Add Topology Error: add instance %s to node %d error: %s",
				instance.InstanceID,
				instance.NodeIndex,
				err.Error())
			logrus.Error(errMsg)
		}
	}, instanceList, 512)
	wg.Wait()
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, resp)
}

func ResetStatusHandler(ctx *gin.Context) {
	nodes, err := synchronizer.GetNodeList()
	if err != nil {
		errMsg := fmt.Sprintf("Get Node List Error: %s", err.Error())
		logrus.Errorf(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	for _, node := range nodes {
		links, err := synchronizer.GetLinkList(node.NodeIndex)
		if err != nil {
			errMsg := fmt.Sprintf("Get Link List of Node %d Error: %s", node.NodeIndex, err.Error())
			logrus.Errorf(errMsg)
		}
		for _, linkInfo := range links {
			err := synchronizer.RemoveLink(node.NodeIndex, linkInfo.GetLinkID())
			if err != nil {
				errMsg := fmt.Sprintf("Remove Link %s of Node %d Error: %s", linkInfo.GetLinkID(), node.NodeIndex, err.Error())
				logrus.Errorf(errMsg)
			}
		}
		instances, err := synchronizer.GetInstanceList(node.NodeIndex)
		if err != nil {
			errMsg := fmt.Sprintf("Get Instance List of Node %d Error: %s", node.NodeIndex, err.Error())
			logrus.Errorf(errMsg)
		}
		for _, instanceInfo := range instances {
			err := synchronizer.RemoveInstance(node.NodeIndex, instanceInfo.InstanceID)
			if err != nil {
				errMsg := fmt.Sprintf("Remove Instance %s of Node %d Error: %s", instanceInfo.InstanceID, node.NodeIndex, err.Error())
				logrus.Errorf(errMsg)
			}
		}
		synchronizer.UpdateNode(node.NodeIndex, func(n *model.Node) error {
			n.FreeInstance += len(instances)
			return nil
		})
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, resp)
}
