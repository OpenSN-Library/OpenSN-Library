import {message} from "antd";
import param from '../parameters.json';

let webSockets = [];
const ws_url = param.ws_url;
// 心跳定时器
let heartbeat_timer = null;
// 心跳发送频率
const heartbeat_interval = 5000;
// 是否自动重连
let is_reconnect = false;
// 重连次数
const reconnect_count = 5;
// 已发起重连次数
let reconnect_current = 1;
// 重连定时器
let reconnect_timer = null;
// 重连频率
const reconnect_interval = 2000;

const init = (port) => {
  /**
   * 初始化webSocket连接
   */
  console.log("初始化websocket连接，端口:", port);
  if (typeof WebSocket === "undefined" || WebSocket === null) {
    message.error("您的浏览器不支持 WebSocket!").then();
    return null;
  } else if (webSockets[port]) {
    return webSockets[port];
  }

  const webSocket = new WebSocket(ws_url + ':' + port);
  webSockets[port] = webSocket;

  webSocket.onmessage = function (event) {
    send(port, "ack");
    document.getElementById('videoPlayer' + port).src = 'data:image/jpeg;base64,' + event.data;
  };

  webSocket.onopen = function () {
    console.log(port + "连接成功");
    is_reconnect = true;
    // 开启心跳
    heartbeat(port);
  };

  webSocket.onclose = function (event) {
    console.log(port + "连接断开 (" + event.code + ")");
    // 清除心跳计时器
    clearInterval(heartbeat_interval);

    // 需要重连
    if (is_reconnect) {
      reconnect_current = 0;
      reconnect_timer = setTimeout(() => {
        // 超过重连次数
        if (reconnect_current >= reconnect_count) {
          clearTimeout(reconnect_timer);
          return;
        }
        reconnect_current++;
        reconnect(port);
      }, reconnect_interval);
    }
  };

  webSocket.onerror = function (event) {
    console.log(port + "连接错误: " + event);
  };

  return webSocket;
}

const send = (port, data, callback = null) => {
  /**
   * 发送消息
   * @param {*} data 发送数据
   * @param {*} callback 发送后的自定义回调函数
   */
  if (webSockets[port].readyState === webSockets[port].OPEN) {
    // wenSocket开启状态直接发送
    webSockets[port].send(JSON.stringify(data))
    if (callback) {
      callback()
    }
  } else if (webSockets[port].readyState === webSockets[port].CONNECTING) {
    // 正在开启状态，则等待1s后重新调用
    setTimeout(function () {
      send(port, data, callback)
    }, 1000)
  } else {
    console.log("websocket未连接!请连接后重新发送。")
  }
};

const heartbeat = (port) => {
  /**
   * 心跳，保持连接状态
   */
  if (heartbeat_timer) {
    clearInterval(heartbeat_timer);
  }

  // heartbeat_timer = setInterval(() => {
  //     send(webSockets[port], {
  //       kind: 0,
  //     });
  // }, heartbeat_interval);
}

const close = (webSocket) => {
  /**
   * 主动关闭连接
   */
  clearInterval(heartbeat_timer);
  is_reconnect = false;
  webSocket.close();
}

const reconnect = (port) => {
  /**
   * 重连
   */
  webSockets[port].close();
  init(port);
}

export const socket = {
  init,
  send,
  close,
}