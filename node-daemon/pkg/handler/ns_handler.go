package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/arranger"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetNsListHandler(ctx *gin.Context) {
	var list []model.Namespace
	data.NamespaceMapLock.RLock()
	defer data.NamespaceMapLock.RUnlock()
	for _, v := range data.NamespaceMap {
		list = append(list, *v)
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
	namespace := &model.Namespace{
		Name: reqObj.Name,
		NsConfig: model.NamespaceConfig{
			ImageMap:           reqObj.NsConfig.ImageMap,
			ContainerEnvs:      reqObj.NsConfig.ContainerEnvs,
			InterfaceAllocated: []string{},
		},
		InstanceAllocInfo: make(map[int][]string),
		LinkAllocInfo:     make(map[int][]string),
	}
	var instanceArray []model.InstanceConfig
	var linkArray []model.LinkConfig
	for i, v := range reqObj.InstConfigs {
		newInstance := model.InstanceConfig{
			InstanceID:         fmt.Sprintf("%s_%s_%d", namespace.Name, v.Type, i),
			Name:               fmt.Sprintf("%s_%d", v.Type, i),
			Type:               v.Type,
			PositionChangeable: v.PositionChangeable,
			Extra:              v.Extra,
		}
		
		if imageName,ok := namespace.NsConfig.ImageMap[v.Type]; ok {
			newInstance.Image = imageName
		} else {
			errMsg := fmt.Sprintf("Type %s of namespace %s has no image to assign",newInstance.Type,namespace.Name)
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
			LinkID:    fmt.Sprintf("%s_%s_%d", namespace.Name, v.Type, i),
			Type:      v.Type,
			Parameter: v.Parameter,
		}
		newLink.InstanceID[0] = instanceArray[v.InstanceIndex[0]].InstanceID
		instanceArray[v.InstanceIndex[0]].LinkIDs = append(instanceArray[v.InstanceIndex[0]].LinkIDs, newLink.LinkID)
		if v.InstanceIndex[1] != -1 {
			newLink.InstanceID[1] = instanceArray[v.InstanceIndex[1]].InstanceID
			instanceArray[v.InstanceIndex[1]].LinkIDs = append(instanceArray[v.InstanceIndex[1]].LinkIDs, newLink.LinkID)
		}
		linkArray = append(linkArray, newLink)
	}
	namespace.InstanceConfig = instanceArray
	namespace.LinkConfig = linkArray
	namespace.AllocatedInstances = len(instanceArray)
	nsBytes, err := json.Marshal(namespace)
	if err != nil {
		errMsg := fmt.Sprintf("Serialize Namespace %s Object Error: %s", namespace.Name, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
			Data:    nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	namespace.Running = false
	data.NamespaceMapLock.Lock()
	data.NamespaceMap[namespace.Name] = namespace
	data.NamespaceMapLock.Unlock()
	utils.LockKeyWithTimeout(key.NamespacesKey, 6*time.Second)
	setResp := utils.RedisClient.HSet(context.Background(), key.NamespacesKey, map[string]interface{}{
		namespace.Name: string(nsBytes),
	})
	if setResp.Err() != nil {
		data.NamespaceMapLock.Lock()
		delete(data.NamespaceMap, namespace.Name)
		data.NamespaceMapLock.Unlock()
		errMsg := fmt.Sprintf("Upload Namespace %s Object Bytes to RedisError: %s", namespace.Name, setResp.Err().Error())
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
	name := ctx.Param("name")

	ns, ok := data.NamespaceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("Unable to Find Namespace by Name: %s", name)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	if ns.Running {
		errMsg := fmt.Sprintf("Namespace %s is Running", name)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	logrus.Infof("Get Config of Namespace %s Success", name)
	targets, err := arranger.ArrangeInstance(ns)
	if err != nil {
		errMsg := fmt.Sprintf("Alloc Instance of namespace %s to nodes error: %s", ns.Name, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	logrus.Infof("Arrange Nodes for namespace %s Success.", name)
	for index, instanceInfos := range targets {
		var list []string
		etcdResp, err := utils.EtcdClient.Get(
			context.Background(),
			fmt.Sprintf(key.NodeInstanceListKeyTemplate, index),
		)
		logrus.Infof("Get Instance List of node %d Success.", index)
		if err != nil {
			errMsg := fmt.Sprintf("Get Instance List of node %d error: %s", index, err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}

		err = json.Unmarshal(etcdResp.Kvs[0].Value, &list)

		if err != nil {
			errMsg := fmt.Sprintf("Parse Instance List of node %d error: %s", index, err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}

		for _, config := range instanceInfos {
			logrus.Infof("Set Instance %s to node %d Success.", config.InstanceID, index)
			list = append(list, config.InstanceID)
			info := model.Instance{
				Config:    config,
				NodeID:    uint32(index),
				Namespace: name,
			}
			infoBytes, err := json.Marshal(info)
			if err != nil {
				errMsg := fmt.Sprintf("Serialize Instance Info of %s error: %s", info.Config.InstanceID, err.Error())
				logrus.Error(errMsg)
				resp := ginmodel.JsonResp{
					Code:    -1,
					Message: errMsg,
				}
				ctx.JSON(http.StatusInternalServerError, resp)
				return
			}
			setResp := utils.RedisClient.HSet(
				context.Background(),
				fmt.Sprintf(key.NodeInstancesKeyTemplate, index),
				[]string{
					info.Config.InstanceID,
					string(infoBytes),
				},
			)
			if setResp.Err() != nil {
				errMsg := fmt.Sprintf("Update Instance Info of %s to Redis error: %s", info.Config.InstanceID, setResp.Err().Error())
				logrus.Error(errMsg)
				resp := ginmodel.JsonResp{
					Code:    -1,
					Message: errMsg,
				}
				ctx.JSON(http.StatusInternalServerError, resp)
				return
			}
		}
		newListStr, err := json.Marshal(list)
		if err != nil {
			errMsg := fmt.Sprintf("Serialize Instance List of Node %d Error: %s", index, err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
		_, err = utils.EtcdClient.Put(
			context.Background(),
			fmt.Sprintf(key.NodeInstanceListKeyTemplate, index),
			string(newListStr),
		)
		if err != nil {
			errMsg := fmt.Sprintf("Update Instance List of Node %d to Etcd Error: %s", index, err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
	}

	for k, array := range targets {
		idArray := make([]string, len(array))
		for index, item := range array {
			idArray[index] = item.InstanceID
		}
		ns.InstanceAllocInfo[k] = idArray
	}

	ns.Running = true
	nsBytes, err := json.Marshal(ns)

	if err != nil {
		errMsg := fmt.Sprintf("Serialize Namespace %s Infomation Error: %s", ns.Name, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	hsetResp := utils.RedisClient.HSet(
		context.Background(),
		key.NamespacesKey,
		ns.Name,
		string(nsBytes),
	)

	if hsetResp.Err() != nil {
		errMsg := fmt.Sprintf("Update Namespace %s Infomation Error: %s", ns.Name, hsetResp.Err().Error())
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

func stopInstances(name string) error {
	data.NamespaceMapLock.Lock()
	defer data.NamespaceMapLock.Unlock()
	info := data.NamespaceMap[name]

	for k, v := range info.InstanceAllocInfo {
		instKey := fmt.Sprintf(key.NodeInstancesKeyTemplate, k)
		delResp := utils.RedisClient.HDel(
			context.Background(),
			instKey,
			v...,
		)
		if delResp.Err() != nil {
			errMsg := fmt.Sprintf("Remove Running Instance of Node %d Error: %s", k, delResp.Err().Error())
			logrus.Error(errMsg)
			return delResp.Err()
		}
		delSet := make(map[string]bool)
		for _, item := range v {
			delSet[item] = true
		}
		var oldList []string
		var newList []string
		listKey := fmt.Sprintf(key.NodeInstanceListKeyTemplate, k)
		getResp, err := utils.EtcdClient.Get(context.Background(), listKey)
		if err != nil {
			logrus.Errorf("Get Etcd Instance List Value of Node %d Error: %s", k, err.Error())
			return err
		}
		if len(getResp.Kvs) <= 0 {
			errMsg := fmt.Sprintf("etcd instance list %s not found", listKey)
			logrus.Errorf("Get Etcd Instance List Value of Node %d Error: %s", k, errMsg)
			return errors.New(errMsg)
		}
		err = json.Unmarshal(getResp.Kvs[0].Value, &oldList)
		if err != nil {
			logrus.Errorf("Parse Instance List of Node %d Error: %s", k, err.Error())
			return err
		}

		for _, item := range oldList {
			if !delSet[item] {
				newList = append(newList, item)
			}
		}
		newListBytes, err := json.Marshal(newList)

		if err != nil {
			logrus.Errorf("Serialize Instance List of Node %d Error: %s", k, err.Error())
			return err
		}
		_, err = utils.EtcdClient.Put(context.Background(), listKey, string(newListBytes))

		if err != nil {
			logrus.Errorf("Update Etcd Instance List of Node %d Error: %s", k, err.Error())
			return err
		}
	}
	info.InstanceAllocInfo = map[int][]string{}
	return nil
}

func removeLinks(name string) error {
	data.NamespaceMapLock.Lock()
	defer data.NamespaceMapLock.Unlock()
	_ = data.NamespaceMap[name]
	return nil
}

func StopNsHandler(ctx *gin.Context) {
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
	if !info.Running {
		errMsg := fmt.Sprintf("Namespace %s is not Running", name)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	err := stopInstances(name)

	if err != nil {
		errMsg := fmt.Sprintf("Remove Instance Error: %s", err.Error())
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	err = removeLinks(name)

	if err != nil {
		errMsg := fmt.Sprintf("Remove Link Error: %s", err.Error())
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	info.Running = false
	infoBytes, err := json.Marshal(info)
	if err != nil {
		errMsg := fmt.Sprintf("Serialize Namespace Info Error: %s", err.Error())
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	state := utils.LockKeyWithTimeout(key.NamespacesKey, 30*time.Second)
	if state {
		setResp := utils.RedisClient.HSet(context.Background(), key.NamespacesKey, name, string(infoBytes))
		if setResp.Err() != nil {
			errMsg := fmt.Sprintf("Remove Instance Error: %s", err.Error())
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
	}

	info.Running = false
	nsBytes, err := json.Marshal(info)

	if err != nil {
		errMsg := fmt.Sprintf("Serialize Namespace %s Infomation Error: %s", info.Name, err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	hsetResp := utils.RedisClient.HSet(
		context.Background(),
		key.NamespacesKey,
		info.Name,
		string(nsBytes),
	)

	if hsetResp.Err() != nil {
		errMsg := fmt.Sprintf("Update Namespace %s Infomation Error: %s", info.Name, hsetResp.Err().Error())
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
}
