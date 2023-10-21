import {Component} from "react";
import "./style.css";
import {Button, message, Select} from "antd";

class TransmitFile extends Component {

  constructor(props) {
    super(props);
    this.state = {
      src_id: "请选择源卫星ID",
      dst_id: "请选择目的卫星ID",
      size: "请输入数据包大小",
      loading: false,
    }
  }

  onChangeSrcId = (value) => {
    this.setState({src_id: value});
  }

  onChangeDstId = (value) => {
    this.setState({dst_id: value});
  }

  onChangeSize = (value) => {
    this.setState({size: value});
  }

  onStartTransmit = () => {
    if (this.state.src_id.startsWith("请")) {
      message.error("请选择源卫星ID！").then();
    } else if (this.state.dst_id.startsWith("请")) {
      message.error("请选择目的卫星ID！").then();
    } else if (this.state.size.startsWith("请")) {
      message.error("请输入数据包大小！").then();
    } else {
      this.setState({loading: true});
      setTimeout(() => {
        this.setState({loading: false});
      }, 5000);
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
    let items_size = [];
    for (let i = 1; i <= 1024; i++) {
      items_size.push({
        value: i + "MB",
        label: i + "MB"
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
                  defaultValue={this.state.dst_id}
                  optionFilterProp={"children"}
                  onChange={this.onChangeDstId}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_id}/>
          <Select className={"route_select"}
                  showSearch={true}
                  defaultValue={this.state.size}
                  optionFilterProp={"children"}
                  onChange={this.onChangeSize}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_size}/>
          <Button className={"route_button_display"} loading={this.state.loading}
                  onClick={this.onStartTransmit}>
            开始传输
          </Button>

        </div>
    )
  }
}

export default TransmitFile;

