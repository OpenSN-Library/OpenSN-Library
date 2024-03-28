import {React,useState} from 'react';
import {Layout, Menu, theme } from 'antd';
import { MdOutlineSatelliteAlt } from "react-icons/md";
import { CreateNamespaceForm } from '../Component/CreateNamespaceForm';
import { NamespaceList } from '../Component/GetNamespaceList';
import { NodeList } from './NodePage';
import { ConsoleItems } from './OverviewPage';
import { FileItems } from './FilePage';
import { InstanceListPage } from './InstancePage';
import { LinkListPage } from './LinkPage';
import { Overview } from './OverviewPage';
import ExportOutlined from '@ant-design/icons/ExportOutlined';
import {Ion} from 'cesium';
import { MonitorPage } from './MonitorPage';
import { DatabasePage } from './DatabasePage';
const { Header, Content, Sider } = Layout;

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIyZWFlNTU1MS1mOGE1LTRiZWEtODc0Zi05NTQ2NDc3Y2MyOWMiLCJpZCI6MTkxNzA5LCJpYXQiOjE3MDYxMDExOTR9.t-UUQ5k6vHbnAXbaF88oB5k0vCEROqeXbGgOktXk9xM
// Ion.defaultAccessToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIyMTg2ODQwZC00NTJiLTQ0MmUtOTM1Mi0yYTk3YTE4OGVlNGMiLCJpZCI6MTkxNzA5LCJpYXQiOjE3MTEyODE2NDR9.1dM1b_tT7qiFajlKeK4kLUFAvYrBMWq4-DlenYtUtJs"

const topBarItems = [
    {
        key:"overview",
        label: `概览`,
    },
    {
      key:"instances",
      label: `节点`,
    },
    {
      key:"links",
      label: `链路`,
    },
    {
        key:"mount_manager",
        label: `文件管理`,
    },
    {
      key:"cluster",
      label: `部署机器`,
    },
    {
      key:"monitor",
      label: `监控`,
    },
    {
      key:"database",
      label: `查看数据库`,
    },
    {
        key:"help",
        label: (<a href="/help" target="_blank">
                  帮助文档<ExportOutlined />
                </a>),
    },
    {
        key:"about",
        label: (<a href="/about" target="_blank">
                  关于<ExportOutlined />
                </a>),
    }
];

const componentPage = {
    "overview":<Overview/>,
    "instances":<InstanceListPage/>,
    "links":<LinkListPage/>,
    "mount_manager":<div/>,
    "cluster":<NodeList/>,
    "monitor":<MonitorPage/>,
    "database":<DatabasePage/>,
}

const App = () => {
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();
  var [selectedTopBarItem,setSelectedTopBarItem] = useState("overview");
  return (
    <Layout style={{height: '100%'}}> 
      <Header
        style={{
          display: 'flex',
          alignItems: 'left',
          paddingLeft: 0
        }}
      >
        <MdOutlineSatelliteAlt style={{color:"white",height:"40px",width:"40px",marginLeft:"80px",marginRight:"80px",marginTop:"12px"}}/>
        
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={selectedTopBarItem}
          items={topBarItems}
          style={{flex: 1, minWidth: 0}}
          onSelect={({key}) => {
            console.log(key)
            if (Object.keys(componentPage).indexOf(key) !== -1){
              console.log(key)
              setSelectedTopBarItem(key);
            }
          }}
        />
      </Header>
      <Layout>
        <Layout
          style={{
            padding: '0 24px 24px',
          }}
        >
          <Content
            style={{
              padding: 24,
              margin: 0,
              minHeight: 280,
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
              overflow: 'scroll'
            }}
          >
            {componentPage[selectedTopBarItem]}
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};
export default App;