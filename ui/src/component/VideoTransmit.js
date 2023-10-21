import {Component} from "react";
import "./style.css"
import {Button, message, Select} from "antd";
import {getVideoTransition} from "../axios";
import {sateLocation2xyz} from "../model/util/sateLocation2xyz";

class VideoTransmit extends Component {

  constructor(props) {
    super(props);
    this.state = {
      src_id: "请选择源卫星ID",
      tcp_dst_id: "请选择TCP传输目的卫星ID",
      my_dst_ip: "请选择新协议传输目的卫星ID",
    }
  }

  onChangeSrcId = (value) => {
    this.setState({src_id: value});
  }

  onChangeTcpDstId = (value) => {
    this.setState({tcp_dst_id: value});
  }

  onChangeMyDstId = (value) => {
    this.setState({my_dst_id: value});
  }

  onClickStart = () => {
    if (this.state.src_id.startsWith("请")) {
      message.error("请选择源卫星ID！").then();
    } else if (this.state.tcp_dst_id.startsWith("请")) {
      message.error("请选择TCP传输目的卫星ID！").then();
    } else if (this.state.my_dst_id.startsWith("请")) {
      message.error("请选择新协议传输目的卫星ID！").then();
    } else {
      getVideoTransition(this.state.src_id, this.state.tcp_dst_id, this.state.my_dst_id);
      const ports = {};
      ports['src_port'] = 30000 + parseInt(this.state.src_id.split('_')[1]);
      ports['tcp_port'] = 30000 + parseInt(this.state.tcp_dst_id.split('_')[1]);
      ports['my_port'] = 30000 + parseInt(this.state.my_dst_id.split('_')[1]);

      setTimeout(() => {
        this.props.startTransmit(ports);
      }, 2000);
    }
  }

  onClickEnd = () => {
    this.props.endTransmit();
  }

  setPlayerPosition = () => {
    const videoPlayerSrc = document.getElementById('videoPlayer' + this.state.ports['src_port']);
    const videoPlayerTcpDst = document.getElementById('videoPlayer' + this.state.ports['tcp_port']);
    const videoPlayerMyDst = document.getElementById('videoPlayer' + this.state.ports['my_port']);
    const srcXYZ = sateLocation2xyz(this.props.getNodePosition(this.state.src_id));
    const tcpDstXYZ = sateLocation2xyz(this.props.getNodePosition(this.state.tcp_dst_id));
    const myDstXYZ = sateLocation2xyz(this.props.getNodePosition(this.state.my_dst_id));

    videoPlayerSrc.style.position = "absolute";
    videoPlayerSrc.style.width = "20%";
    videoPlayerSrc.style.height = "20%";
    videoPlayerSrc.style.right = srcXYZ.x - 50 + "px";
    videoPlayerSrc.style.bottom = srcXYZ.y - 50 + "px";

    videoPlayerTcpDst.style.position = "absolute";
    videoPlayerTcpDst.style.width = "20%";
    videoPlayerTcpDst.style.height = "20%";
    videoPlayerTcpDst.style.left = tcpDstXYZ.x + 50 + "px";

    videoPlayerMyDst.style.position = "absolute";
    videoPlayerMyDst.style.width = "20%";
    videoPlayerMyDst.style.height = "20%";
    videoPlayerMyDst.style.left = myDstXYZ.x + 50 + "px";

    if (tcpDstXYZ.y > myDstXYZ.y) {
      videoPlayerTcpDst.style.bottom = tcpDstXYZ.y + "px";
      videoPlayerMyDst.style.top = myDstXYZ.y + "px";
    } else {
      videoPlayerTcpDst.style.top = tcpDstXYZ.y + "px";
      videoPlayerMyDst.style.bottom = myDstXYZ.y + "px";
    }
  }

  render() {
    let para = this.props.para;
    let items_id = [];
    for (let node_id in para["all_node"]) {
      items_id.push({
        value: para["all_node"][node_id],
        label: "卫星 " + para["all_node"][node_id]
      })
    }
    return (
        <div>
          <Select className={"route_select"}
                  showSearch={true}
                  defaultValue={this.state.src_id}
                  optionFilterProp={"children"}
                  onChange={this.onChangeSrcId}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_id}/>
          <Select className={"route_select"}
                  showSearch={true}
                  defaultValue={this.state.tcp_dst_id}
                  optionFilterProp={"children"}
                  onChange={this.onChangeTcpDstId}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_id}/>
          <Select className={"route_select"}
                  showSearch={true}
                  defaultValue={this.state.my_dst_ip}
                  optionFilterProp={"children"}
                  onChange={this.onChangeMyDstId}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_id}/>
          <Button className={"video_transmit_start"} onClick={this.onClickStart}>
            开始传输
          </Button>
          <Button className={"video_transmit_end"} onClick={this.onClickEnd}>
            结束传输
          </Button>
        </div>
    )
  }
}

export default VideoTransmit;