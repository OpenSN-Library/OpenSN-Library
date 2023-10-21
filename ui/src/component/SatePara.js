import {Component} from "react";
import "./style.css"
import {Card, Select, Space} from "antd";

class SatePara extends Component {

  onChange = (value) => {
    this.props.setNodeId(value);
  }

  render() {
    let para = this.props.getSatePara(this.props.node_id);
    let items = [];
    for (let node_id in para["all_node"]) {
      items.push({
        value: para["all_node"][node_id],
        label: "卫星: " + para["all_node"][node_id]
      })
    }
    return (
        <Space directions={"vertical"} size={16}>
          <Select className={"sate_select"}
                  showSearch={true}
                  defaultValue={this.props.node_id}
                  optionFilterProp={"children"}
                  onChange={this.onChange}
                  filterOption={(input, option) =>
                      (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={items}/>
          <Card className={"sate_location_card"}>
            <h2 style={{marginTop: "-20px"}}>{"\u00A0" + this.props.node_id}</h2>
            <Card style={{marginTop: "-10px"}} size={"small"} title={"位置信息"}>
              <p style={{marginTop: "-5px"}}>{"经度: " + (para["location"]["long"] / Math.PI * 180).toFixed(6) + "°"}</p>
              <p style={{marginTop: "-5px"}}>{"纬度: " + (para["location"]["lat"] / Math.PI * 180).toFixed(6) + "°"}</p>
              <p style={{
                marginTop: "-5px",
                marginBottom: "-5px"
              }}>{"高度: " + (para["location"]["height"] / 1000).toFixed(5) + " km"}</p>
            </Card>
            <Card size={"small"} title={"连接信息"}>
              <div style={{marginBottom: "-10px"}}>{
                Object.keys(para.connections).map((value) => {
                  let connection = para.connections[value];
                  return <p className={"sate_connect_p"}>
                    {"端口IP:\u00A0\u00A0" + connection["ip"]}
                    <br/>
                    {"目的卫星:\u00A0\u00A0" + connection["dst_id"]}
                  </p>;
                })
              }</div>
            </Card>
          </Card>
        </Space>
    )
  }
}

export default SatePara;