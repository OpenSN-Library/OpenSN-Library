import {Component} from "react";
import {Switch} from "antd";

class LinkSet extends Component {

  onChangeConnect = (checked: boolean) => {
    this.props.flag_change("connect", checked);
  }

  onChangeLabel = (checked: boolean) => {
    this.props.flag_change("label", checked);
  }

  onChangeLongLine = (checked: boolean) => {
    this.props.flag_change("longLine", checked);
  }

  onChangeSateUpdate = (checked: boolean) => {
    this.props.flag_change("sateUpdate", checked);
  }

  render() {
    return (
        <div>
          <div className={"linkSet_div"}>
            卫星连接线显示
            <Switch className={"linkSet_switch_connect"}
                    defaultChecked onChange={this.onChangeConnect}/>
          </div>
          <div className={"linkSet_div"}>
            卫星标签显示
            <Switch className={"linkSet_switch"}
                    defaultChecked onChange={this.onChangeLabel}/>
          </div>
          <div className={"linkSet_div"}>
            卫星轨道显示
            <Switch className={"linkSet_switch"}
                    defaultChecked onChange={this.onChangeLongLine}/>
          </div>
          <div className={"linkSet_div"}>
            卫星位置更新
            <Switch className={"linkSet_switch"}
                    defaultChecked onChange={this.onChangeSateUpdate}/>
          </div>
        </div>
    )
  }
}

export default LinkSet;