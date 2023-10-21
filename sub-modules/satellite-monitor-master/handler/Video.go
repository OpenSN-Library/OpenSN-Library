package handler

import (
	"net/http"
	"satellite/monitor/model"
	"satellite/monitor/utils"

	"github.com/gin-gonic/gin"
)

func TransitionVideo(ctx *gin.Context) {
	var req = model.TransitionVideoReq {
		SrcID: ctx.Query("src_id"),
		TcpDstID: ctx.Query("tcp_dst_id"),
		MyDstID: ctx.Query("my_dst_id"),
	}
	var err error
	var resp model.JsonResp
	
	err = utils.StartCompress(req.MyDstID)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError,resp)
		return
	}
	err = utils.StartCompress(req.TcpDstID)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError,resp)
		return
	}
	err = utils.StartRecv(req.MyDstID)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError,resp)
		return
	}
	err = utils.StartRecv(req.TcpDstID)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError,resp)
		return
	}
	err = utils.StartSend(req.SrcID,req.TcpDstID,req.MyDstID)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError,resp)
		return
	}
	resp.Code = 0
	resp.Message = "success"
	ctx.JSON(http.StatusOK,resp)
	return

}