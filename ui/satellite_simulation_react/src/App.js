import React, {useEffect, useState} from "react";
import './App.css'

import Model from './model/model.js'
import LinkSet from "./component/LinkSet";
import SatePara from "./component/SatePara";
import RoutePath from "./component/RoutePath";
import TransmitFile from "./component/TransmitFile";
import logo from './img/logo.png'

import {Layout, Menu} from "antd";
import Sider from "antd/es/layout/Sider";
import {
  ApartmentOutlined,
  ApiOutlined,
  BarChartOutlined,
  BlockOutlined,
  VideoCameraOutlined,
} from "@ant-design/icons"
import {Content} from "antd/es/layout/layout";
import para from "./parameters.json";

import {
  getSatelliteStatus,
  getSatelliteList,
  getSatelliteInterfaces,
  getCommandTraceroute, getGroundPosition, getGroundConnection,
} from "./axios";
import VideoTransmit from "./component/VideoTransmit";
import VideoTransmitPage from "./pages/VideoTransmitPage";


const items = [
  {
    key: '1',
    icon: <BarChartOutlined/>,
    label: '卫星参数',
  },
  {
    key: '2',
    icon: <ApartmentOutlined/>,
    label: '路由路径',
  },
  {
    key: '4',
    icon: <BlockOutlined/>,
    label: '文件传输',
  },
  {
    key: '3',
    icon: <ApiOutlined/>,
    label: '连接设置',
  },
  {
    key: '5',
    icon: <VideoCameraOutlined/>,
    label: '视频传输',
  },
];

const App = () => {
  const [sateLocations, setSateLocations] = useState(undefined);
  const [sateConnections, setSateConnections] = useState(undefined);
  const [groundPosition, setGroundPosition] = useState(undefined);  // [经度，纬度]
  const [groundConnections, setGroundConnections] = useState(undefined);
  const [interfaces, setInterfaces] = useState([]);
  const [routePath, setRoutePath] = useState([]);
  const [sateParaNodeId, setSateParaNodeId] = useState(para.satellite_name + "0");  // 默认选择的卫星是Sat_0
  const [key, setKey] = useState('');  // 选择的菜单项
  const [videoTransmitPara, setVideoTransmitPara] = useState({flag:false, ports:{}});//视频传输参数
  const [flags, setFlags] = useState({
    "connect": true,
    "longLine": true,
    "label": true,
    "sateUpdate": true,
  });


  useEffect(() => {
    getSatelliteList((data) => {
      setSateConnections(data);
    });
    getSatelliteInterfaces((data) => {
      setInterfaces(data);
    });
    getGroundPosition((data) => {
      setGroundPosition(data);
    });

    // 添加定时器，每秒获取一次卫星位置，地面站连接情况
    const interval = setInterval(() => {
      if (!flags["sateUpdate"]) {
        return;
      }
      getSatelliteStatus((data) => {
        setSateLocations(data);
      });
      getGroundConnection((data) => {
        setGroundConnections(data);
      });
    }, 1000);

    // 页面关闭时删除定时器
    return () => {
      clearInterval(interval);
    }
  }, [flags]);

  const getSatePara = (node_id) => {
    let all_node = [];
    for (let key in sateLocations) {
      all_node.push(key);
    }
    return {
      "all_node": all_node,
      "location": sateLocations[node_id],
      "connections": interfaces[node_id],
    }
  }

  const getAllRoute = () => {
    let all_node = [];
    for (let key in sateLocations) {
      all_node.push(key);
    }
    return {
      "all_node": all_node,
      "interfaces": interfaces,
    }
  }

  const getRoutePath = (src_id, dst_ip) => {
    getCommandTraceroute(src_id, dst_ip, (data) => {
      setRoutePath(data);
    });
  }

  return (
      <Layout style={{
        height: '100%',
        width: '100%',
      }}>
        <Sider width={260} trigger={null} collapsible collapsed={false}>
          <img className={"logo"} src={logo} alt={"logo"}/>
          <Menu className={"menu"}
                theme="dark"
                mode="inline"
                defaultSelectedKeys={key}
                items={items}
                onClick={(item) => {
                  if (key === item.key) {
                    setKey("");
                  } else {
                    setKey(item.key);
                  }
                }}
          />
          <div className="data_table">
            {
              (key === '1') ?
                  <SatePara getSatePara={getSatePara} node_id={sateParaNodeId}
                            setNodeId={(node_id) => {
                              setSateParaNodeId(node_id);
                            }}/> :
              (key === '2') ?
                  <RoutePath para={getAllRoute()} routePath={routePath} getRoutePath={getRoutePath}
                             clear={() => setRoutePath([])}/> :
              (key === '3') ? <LinkSet flags={flags}
                                       flag_change={(str, bool) => {
                                         let tmp = flags;
                                         tmp[str] = bool;
                                         setFlags(tmp);
                                       }}/> :
              (key === '4') ? <TransmitFile para={getAllRoute()}/> :
              (key === '5') ? <VideoTransmit para={getAllRoute()}
                                             startTransmit={(ports) => {
                                               setVideoTransmitPara({flag:true, ports:ports});
                                               console.log(videoTransmitPara);
                                             }}
                                             endTransmit={() =>{
                                               setVideoTransmitPara({flag:false, ports:{}});
                                               console.log(videoTransmitPara);
                                             }}/> :
              undefined
            }
          </div>
        </Sider>
        <div>
          {
            (videoTransmitPara.flag && key === '5') ? <VideoTransmitPage ports={videoTransmitPara.ports}/> : undefined
          }
        </div>
        <Layout className="site-layout">
          <Content style={{
            height: '100%',
            width: '100%',
            backgroundColor: '#1d1d1d'
          }}>
            <Model sateLocations={sateLocations}
                   sateConnections={sateConnections}
                   groundPosition={groundPosition}
                   groundConnections={groundConnections}
                   routePath={routePath}
                   selected_sate={sateParaNodeId}
                   flags={flags}
                   displayPara={(node_id) => {
                     setSateParaNodeId(node_id);
                     setKey("1");
                   }}/>
          </Content>
        </Layout>
      </Layout>
  )
}

export default App;