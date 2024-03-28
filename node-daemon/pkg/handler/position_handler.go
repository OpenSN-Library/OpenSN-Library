package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetInstancePositionHandler(ctx *gin.Context) {
	instanceID := ctx.Param("instance_id")
	var respData map[string]model.Position
	var err error
	if instanceID == "" || instanceID == "all" {
		respData, err = synchronizer.GetAllInstancePosition()
		if err != nil {
			errMsg := fmt.Sprintf("Get all instance position error: %s", err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
	} else {
		position, err := synchronizer.GetInstancePosition(instanceID)
		if err != nil {
			errMsg := fmt.Sprintf("Get instance %s position error: %s", instanceID, err.Error())
			logrus.Error(errMsg)
			resp := ginmodel.JsonResp{
				Code:    -1,
				Message: errMsg,
				Data:    nil,
			}
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
		respData = map[string]model.Position{
			instanceID: position,
		}
	}

	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, resp)
}
