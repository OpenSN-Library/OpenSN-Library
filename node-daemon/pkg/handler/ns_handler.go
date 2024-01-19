package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/arranger"
	"NodeDaemon/pkg/link"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetNsListHandler(ctx *gin.Context) {
	var list []ginmodel.NamespaceAbstract
	data.NamespaceMapLock.RLock()
	defer data.NamespaceMapLock.RUnlock()
	for _, v := range data.NamespaceMap {
		newAbstract := ginmodel.NamespaceAbstract{
			Name:        v.Name,
			InstanceNum: len(v.InstanceConfig),
			LinkNum:     len(v.LinkConfig),
			Running:     v.Running,
			
		}
		if v.Running {
			newAbstract.AllocNodeIndex = utils.MapKeys[int, []string](v.InstanceAllocInfo)
		}
		list = append(list, newAbstract)
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    list,
	}

	ctx.JSON(http.StatusOK, resp)
}

func GetNsInfoHandler(ctx *gin.Context) {
	name := ctx.Param("name")
	info, ok := data.NamespaceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("Namespace %s Not Found", name)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusNotFound, resp)
	}
	infoData := ginmodel.NamespaceInfoData{
		Name: info.Name,
	}
	ctx.JSON(http.StatusOK, infoData)
}

func CreateNsHandler(ctx *gin.Context) {
	var reqObj ginmodel.CreateNamespaceReq
	err := ctx.Bind(&reqObj)
	if err != nil {
		errMsg := fmt.Sprintf("Parse Create Namespace Request Object Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		})
		return
	}

	if len(reqObj.Name) > 5 {
		err := fmt.Errorf("namespace length cannot be more than 5 character")
		errMsg := fmt.Sprintf("Parse Create Namespace Request Object Error: %s", err.Error())
		logrus.Error(errMsg)
		ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		})
		return
	}

	namespace := &model.Namespace{
		Name: reqObj.Name,
		NsConfig: model.NamespaceConfig{
			ImageMap:           reqObj.NsConfig.ImageMap,
			ContainerEnvs:      reqObj.NsConfig.ContainerEnvs,
			InterfaceAllocated: []string{},
			ResourceLimitMap:   make(map[string]model.ResourceLimit),
		},
		InstanceAllocInfo: make(map[int][]string),
		LinkAllocInfo:     make(map[int][]string),
	}

	for k, v := range reqObj.NsConfig.ResourceMap {
		cpuLimit, err := utils.ParseDecNumber(v.NanoCPU)
		if err != nil {
			errMsg := fmt.Sprintf("Parse Create Namespace Request Object Error: %s", err.Error())
			logrus.Error(errMsg)
			ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			})
			return
		}
		memLimit, err := utils.ParseBinNumber(v.MemoryByte)
		if err != nil {
			errMsg := fmt.Sprintf("Parse Create Namespace Request Object Error: %s", err.Error())
			logrus.Error(errMsg)
			ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			})
			return
		}
		namespace.NsConfig.ResourceLimitMap[k] = model.ResourceLimit{
			NanoCPU:    cpuLimit,
			MemoryByte: memLimit,
		}
	}

	var instanceArray []model.InstanceConfig
	var linkArray []model.LinkConfig
	for i, v := range reqObj.InstConfigs {
		newInstance := model.InstanceConfig{
			InstanceID: fmt.Sprintf("%s_%s_%d", namespace.Name, v.Type, i),
			Name:       fmt.Sprintf("%s_%d", v.Type, i),
			Type:       v.Type,
			DeviceInfo: make(map[string]model.DeviceRequireInfo),
			Extra:      v.Extra,
			Resource:   namespace.NsConfig.ResourceLimitMap[v.Type],
		}

		if imageName, ok := namespace.NsConfig.ImageMap[v.Type]; ok {
			newInstance.Image = imageName
		} else {
			errMsg := fmt.Sprintf("Type %s of namespace %s has no image to assign", newInstance.Type, namespace.Name)
			logrus.Error(errMsg)
			ctx.JSON(http.StatusBadRequest, ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			})
			return
		}
		instanceArray = append(instanceArray, newInstance)
	}
	for i, v := range reqObj.LinkConfigs {
		newLink := model.LinkConfig{
			LinkID:        fmt.Sprintf("%s_%s%d", namespace.Name, v.Type[0:1], i), // Max Len:12
			Type:          v.Type,
			InitParameter: v.Parameter,
		}
		if v.InstanceIndex[0] >= 0 {
			newLink.InitEndInfos[0] = model.EndInfoType{
				InstanceID:   instanceArray[v.InstanceIndex[0]].InstanceID,
				InstanceType: instanceArray[v.InstanceIndex[0]].Type,
			}
			instanceArray[v.InstanceIndex[0]].LinkIDs = append(instanceArray[v.InstanceIndex[0]].LinkIDs, newLink.LinkID)
		}
		if v.InstanceIndex[1] >= 0 {
			newLink.InitEndInfos[1] = model.EndInfoType{
				InstanceID:   instanceArray[v.InstanceIndex[1]].InstanceID,
				InstanceType: instanceArray[v.InstanceIndex[1]].Type,
			}
			instanceArray[v.InstanceIndex[1]].LinkIDs = append(instanceArray[v.InstanceIndex[1]].LinkIDs, newLink.LinkID)

		}
		if deviceInfo, ok := link.LinkDeviceInfoMap[newLink.Type]; ok {
			instanceArray[v.InstanceIndex[0]].DeviceInfo[newLink.Type] = deviceInfo[0]
			if v.InstanceIndex[1] >= 0 {
				instanceArray[v.InstanceIndex[1]].DeviceInfo[newLink.Type] = deviceInfo[1]
			}
		} else {
			logrus.Errorf("Unsupport Link Device Type: %s", newLink.Type)
		}
		linkArray = append(linkArray, newLink)
	}

	namespace.InstanceConfig = instanceArray
	namespace.LinkConfig = linkArray
	namespace.AllocatedInstances = len(instanceArray)
	arranger.ArrangeV4Addr(namespace, 30)
	// arranger.ArrangeV6Addr(namespace, 62)
	namespace.Running = false
	data.NamespaceMapLock.Lock()
	data.NamespaceMap[namespace.Name] = namespace
	data.NamespaceMapLock.Unlock()
	err = synchronizer.PostNamespaceInfo(namespace)
	if err != nil {
		data.NamespaceMapLock.Lock()
		delete(data.NamespaceMap, namespace.Name)
		data.NamespaceMapLock.Unlock()
		errMsg := fmt.Sprintf("Upload Namespace %s Object Bytes to RedisError: %s", namespace.Name, err.Error())
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
	}
	ctx.JSON(http.StatusOK, resp)
}

func UpdateNsHandler(ctx *gin.Context) {
	var req ginmodel.UpdateNamespaceReq
	name := ctx.Param("name")
	info, ok := data.NamespaceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("Namespace %s Not Found", name)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusNotFound, resp)
	}
	if info.Running {
		errMsg := fmt.Sprintf("Namespace %s is Running", name)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
	}
	err := ctx.Bind(req)
	if err != nil {
		errMsg := fmt.Sprintf("Parse Request Data Error : %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	info.NsConfig.ImageMap = req.NsConfig.ImageMap
	info.NsConfig.ContainerEnvs = req.NsConfig.ContainerEnvs
	info.InstanceConfig = make([]model.InstanceConfig, len(req.InstConfigs))
	info.LinkConfig = make([]model.LinkConfig, len(req.LinkConfigs))
	info.AllocatedInstances = len(req.InstConfigs)

}

func StartNsHandler(ctx *gin.Context) {
	var errMsg string
	var resp ginmodel.JsonResp

	goto Start

	// Error Handle Code Segment Start
Error500:
	resp.Code = -1
	resp.Message = errMsg
	logrus.Error(errMsg)
	ctx.JSON(http.StatusInternalServerError, resp)
	return
Error400:
	resp.Code = -1
	resp.Message = errMsg
	logrus.Error(errMsg)
	ctx.JSON(http.StatusBadRequest, resp)
	return
OK200:
	resp.Code = 0
	resp.Message = "Success"
	ctx.JSON(http.StatusOK, resp)
	return
	// Error Handle Code Segment End

Start:

	name := ctx.Param("name")

	ns, ok := data.NamespaceMap[name]
	if !ok {
		errMsg = fmt.Sprintf("Unable to Find Namespace by Name: %s", name)
		goto Error400
	}
	if ns.Running {
		errMsg = fmt.Sprintf("Namespace %s is Running", name)
		goto Error400
	}

	instanceTarget, err := arranger.ArrangeInstance(ns)
	if err != nil {
		errMsg = fmt.Sprintf("Allocate Instance of namespace %s to nodes error: %s", ns.Name, err.Error())
		goto Error500
	}

	for index, instanceInfos := range instanceTarget {

		list, err := synchronizer.GetNodeInstanceList(index)

		if err != nil {
			errMsg = fmt.Sprintf("Get Instance List of node %d error: %s", index, err.Error())
			goto Error500
		}

		list, err = synchronizer.AddInstanceInfosToNode(index, instanceInfos, name, list)

		if err != nil {
			errMsg = fmt.Sprintf("Add InstanceInfos To Node %d error: %s", index, err.Error())
			goto Error500
		}

		err = synchronizer.PostNodeInstanceList(index, list)
		if err != nil {
			errMsg = fmt.Sprintf("Update Instance List to Node %d Error: %s", index, err.Error())
			goto Error500
		}
	}

	linkTarget, err := arranger.ArrangeLink(ns, instanceTarget)

	if err != nil {
		errMsg = fmt.Sprintf("Alloc Link of namespace %s to nodes error: %s", ns.Name, err.Error())
		goto Error500
	}

	for index, linkInfos := range linkTarget {

		list, err := synchronizer.GetNodeLinkList(index)

		if err != nil {
			errMsg = fmt.Sprintf("Get Link List of node %d error: %s", index, err.Error())
			goto Error500
		}

		list, err = synchronizer.AddLinkInfosToNode(index, linkInfos, name, list)

		if err != nil {
			errMsg = fmt.Sprintf("Add Link Info To Node %d error: %s", index, err.Error())
			goto Error500
		}

		err = synchronizer.PostNodeLinkList(index, list)
		if err != nil {
			errMsg = fmt.Sprintf("Update Link List to Node %d Error: %s", index, err.Error())
			goto Error500
		}
	}

	for k, array := range instanceTarget {
		ns.InstanceAllocInfo[k] = utils.SliceMap[*model.InstanceConfig, string](
			func(i *model.InstanceConfig) string {
				return i.InstanceID
			},
			array,
		)
	}

	for k, array := range linkTarget {
		ns.LinkAllocInfo[k] = utils.SliceMap[*model.LinkConfig, string](
			func(i *model.LinkConfig) string {
				return i.LinkID
			},
			array,
		)
	}

	ns.Running = true

	err = synchronizer.PostNamespaceInfo(ns)

	if err != nil {
		errMsg = fmt.Sprintf("Update Namespace %s Info Error: %s", ns.Name, err.Error())
		goto Error500
	}

	goto OK200
}

func stopInstances(name string) error {
	data.NamespaceMapLock.Lock()
	defer data.NamespaceMapLock.Unlock()
	info := data.NamespaceMap[name]
	for k, v := range info.InstanceAllocInfo {
		list, err := synchronizer.GetNodeInstanceList(k)
		if err != nil {
			logrus.Errorf("Get Etcd Instance List Value of Node %d Error: %s", k, err.Error())
			return err
		}
		list, err = synchronizer.DelInstanceInfosFromNode(k, v, name, list)
		if err != nil {
			logrus.Errorf("Remove Instance Infos of Node %d Error: %s", k, err.Error())
			return err
		}
		err = synchronizer.PostNodeInstanceList(k, list)
		if err != nil {
			logrus.Errorf("Update Etcd Instance List Value of Node %d Error: %s", k, err.Error())
			return err
		}
	}
	info.InstanceAllocInfo = map[int][]string{}
	return nil
}

func removeLinks(name string) error {
	data.NamespaceMapLock.Lock()
	defer data.NamespaceMapLock.Unlock()
	info := data.NamespaceMap[name]
	for k, v := range info.LinkAllocInfo {
		list, err := synchronizer.GetNodeLinkList(k)
		if err != nil {
			logrus.Errorf("Get Etcd Link List Value of Node %d Error: %s", k, err.Error())
			return err
		}
		list, err = synchronizer.DelLinkInfosFromNode(k, v, name, list)
		if err != nil {
			logrus.Errorf("Remove Link Infos of Node %d Error: %s", k, err.Error())
			return err
		}
		err = synchronizer.PostNodeLinkList(k, list)
		if err != nil {
			logrus.Errorf("Update Etcd Link List Value of Node %d Error: %s", k, err.Error())
			return err
		}
	}
	info.LinkAllocInfo = map[int][]string{}
	return nil
}

func StopNsHandler(ctx *gin.Context) {

	var errMsg string
	var resp ginmodel.JsonResp

	goto Start

	// Error Handle Code Segment Start
Error500:
	resp.Code = -1
	resp.Message = errMsg
	logrus.Error(errMsg)
	ctx.JSON(http.StatusInternalServerError, resp)
	return
Error400:
	resp.Code = -1
	resp.Message = errMsg
	logrus.Error(errMsg)
	ctx.JSON(http.StatusBadRequest, resp)
	return
Error404:
	resp.Code = -1
	resp.Message = errMsg
	logrus.Error(errMsg)
	ctx.JSON(http.StatusNotFound, resp)
	return
OK200:
	resp.Code = 0
	resp.Message = "Success"
	ctx.JSON(http.StatusOK, resp)
	return
	// Error Handle Code Segment End

Start:

	name := ctx.Param("name")
	info, ok := data.NamespaceMap[name]
	if !ok {
		errMsg = fmt.Sprintf("Namespace %s Not Found", name)
		goto Error404
	}
	if !info.Running {
		errMsg = fmt.Sprintf("Namespace %s is not Running", name)
		goto Error400
	}

	err := removeLinks(name)
	if err != nil {
		errMsg = fmt.Sprintf("Remove Link Error: %s", err.Error())
		goto Error500
	}

	err = stopInstances(name)

	if err != nil {
		errMsg = fmt.Sprintf("Remove Instance Error: %s", err.Error())
		goto Error500
	}

	info.Running = false
	synchronizer.PostNamespaceInfo(info)

	if err != nil {
		errMsg = fmt.Sprintf("Update Namespace %s Error: %s", info.Name, err.Error())
		goto Error500
	}

	goto OK200
}

func DeleteNsHandler(ctx *gin.Context) {
	name := ctx.Param("name")
	info, ok := data.NamespaceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("Namespace %s Not Found", name)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusNotFound, resp)
	}
	if info.Running {
		errMsg := fmt.Sprintf("Namespace %s is Running", name)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	delResp := utils.RedisClient.HDel(context.Background(), key.NamespacesKey, name)
	if delResp.Err() != nil {
		errMsg := fmt.Sprintf("Delete Namespace %s Info Error:%s", name, delResp.Err().Error())
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
