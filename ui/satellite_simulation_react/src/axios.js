import para from "./parameters.json";
import axios from "axios";
import {message} from "antd";

export const getSatelliteStatus = (callback) => {
  axios.get(para.url + '/api/satellite/status').then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.log("get locations error:" + error);
  })
}

export const getSatelliteList = (callback) => {
  axios.get(para.url + '/api/satellite/list').then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.log("get locations error:" + error);
  })
}

export const getSatelliteInterfaces = (callback) => {
  axios.get(para.url + '/api/satellite/interfaces').then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.log("get interfaces error:" + error);
  })
}

export const getCommandTraceroute = (src_id, dst_ip, callback) => {
  axios.get(para.url + '/api/command/traceroute?src_id=' + src_id + '&dst_ip=' + dst_ip).then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.error("get route path error:" + error);
    message.error("不存在路由路径！").then();
  })
}

export const getGroundPosition = (callback) => {
  axios.get(para.url + '/api/ground/position').then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.log("get ground position error:" + error);
  })
}

export const getGroundConnection = (callback) => {
  axios.get(para.url + '/api/ground/connection').then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
        callback(response.data.data);
      }).catch((error) => {
    console.log("get ground status error:" + error);
  })
}

export const getVideoTransition = (src_id, tcp_dst_id, my_dst_id) => {
  axios.get(para.url + '/api/video/transition?src_id=' + src_id + "&tcp_dst_id=" + tcp_dst_id + "&my_dst_id=" + my_dst_id).then(
      (response) => {
        if (para.console_print) {
          console.log(response.data.data);
        }
      }).catch((error) => {
    console.log("post video transition error:" + error);
  })
}