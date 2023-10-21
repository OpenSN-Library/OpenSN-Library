import {Component} from "react";
import "./style.css"
import {Button, Card, message, Select} from "antd";

class RoutePath extends Component {

  constructor(props) {
    super(props);
    this.state = {
      src_id: "请选择源卫星ID",
      dst_id: "请选择目的卫星ID",
      dst_ip: "请选择目的卫星IP",
      display: false,
    }
  }

  onChangeSrcId = (value) => {
    this.setState({src_id: value});
  }

  onChangeDstId = (value) => {
    this.setState({dst_id: value});
  }

  onChangeDstIp = (value) => {
    this.setState({dst_ip: this.props.para["interfaces"][this.state.dst_id][value]["ip"]});
  }

  onClickDisplay = () => {
    if (this.state.src_id.startsWith("请")) {
      message.error("请选择源卫星ID！").then();
    } else if (this.state.dst_id.startsWith("请")) {
      message.error("请选择目的卫星ID！").then();
    } else if (this.state.dst_ip.startsWith("请")) {
      message.error("请选择目的卫星IP！").then();
    } else {
      this.setState({display: true})
      this.props.getRoutePath(this.state.src_id, this.state.dst_ip);
    }
  }

  onClickClear = () => {
    this.setState({display: false})
    this.props.clear();
  }

  componentWillUnmount() {
    this.onClickClear();
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
    let items_ip = [];
    let interfaces = this.props.para["interfaces"][this.state.dst_id]
    for (let node_ip in interfaces) {
      items_ip.push({
        value: node_ip,
        label: interfaces[node_ip]["ip"]
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
                  defaultValue={this.state.dst_ip}
                  optionFilterProp={"children"}
                  onChange={this.onChangeDstIp}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items_ip}/>
          <Button className={"route_button_display"} onClick={this.onClickDisplay}>
            显示路径
          </Button>
          <Button className={"route_button_clear"} onClick={this.onClickClear}>
            清除路径
          </Button>
          {
            (this.state.display) ?
                <Card className={"route_path_card"} size={"small"} title={"路由路径"}>
                  <div>{
                    Object.keys(this.props.routePath).map((value) => {
                      return <p>{"时延: " + (this.props.routePath[value]["latency"] * 500).toFixed(3) + "ms ---> " + this.props.routePath[value]["node_id"]}</p>;
                    })
                  }</div>
                </Card> : undefined
          }
        </div>
    )
  }
}

export default RoutePath;