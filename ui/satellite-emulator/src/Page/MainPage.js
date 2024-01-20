import {React,useState} from 'react';
import { Breadcrumb, Layout, Menu, theme } from 'antd';
import { MdOutlineSatelliteAlt } from "react-icons/md";
import { CreateNamespaceForm } from '../Component/CreateNamespaceForm';
import { NamespaceList } from '../Component/GetNamespaceList';
import { NodeList } from '../Component/GetNodeList';
const { Header, Content, Sider } = Layout;


const topBarItems = [
    {
        key:"console",
        label: `控制台`,
    },
    {
        key:"mount_manager",
        label: `文件管理`,
    },
    {
        key:"help",
        label: `使用文档`,
    },
    {
        key:"about",
        label: `关于`,
    }
];

const consoleItems = [
  {
    key: `add_namespace`,
    label: `新建项目`,
    children: null
  },
  {
    key: `namespace_list`,
    label: `项目列表`,
    children: null
  },
  {
    key: `monitor`,
    label: `监控`,
    children: null
  },
  {
    key: `node_list`,
    label: `计算节点列表`,
    children: null
  }
]

const fileItems = [
  {
    key: `file_manage`,
    label: `文件管理`,
    children: null
  }
]

const AboutItem = [
  {
    key: `about`,
    label: `关于`,
    children: null
  }
]

const componentMap = {
    "add_namespace": <CreateNamespaceForm/>,
    "namespace_list": <NamespaceList/>,
    "node_list": <NodeList/>,
}

const siderMemuMap = {
  "console": consoleItems,
  "mount_manager": fileItems,
  "about": AboutItem,
}
const App = () => {
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  var [selectedTopBarItem,setSelectedTopBarItem] = useState("console");
  var [selectedSiderItem,setSelectedSiderItem] = useState("add_namespace");
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
          defaultSelectedKeys={selectedSiderItem}
          items={topBarItems}
          style={{flex: 1, minWidth: 0}}
          onSelect={({key}) => {
            setSelectedTopBarItem(key);
            if (siderMemuMap[key] != null && siderMemuMap[key].length > 0){
              selectedSiderItem = siderMemuMap[key][0].key
              setSelectedSiderItem(selectedSiderItem);
            } else {
              selectedSiderItem = ""
              setSelectedSiderItem(selectedSiderItem);
            }
          }}
        />
      </Header>
      <Layout>
        <Sider
          width={200}
          style={{
            background: colorBgContainer,
          }}
        >
          <Menu
            mode="inline"
            defaultSelectedKeys={selectedSiderItem}
            defaultOpenKeys={['sub1']}
            style={{
              height: '100%',
              borderRight: 0,
            }}
            items={siderMemuMap[selectedTopBarItem]}
            onSelect={({key}) => {
              setSelectedSiderItem(key);
            }}
          />
        </Sider>
        <Layout
          style={{
            padding: '0 24px 24px',
          }}
        >
          <Breadcrumb
            style={{
              margin: '16px 0',
            }}
          >
            <Breadcrumb.Item>Home</Breadcrumb.Item>
            <Breadcrumb.Item>List</Breadcrumb.Item>
            <Breadcrumb.Item>App</Breadcrumb.Item>
          </Breadcrumb>
          <Content
            style={{
              padding: 24,
              margin: 0,
              minHeight: 280,
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
            }}
          >
            {componentMap[selectedSiderItem]}
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};
export default App;