import React from 'react';
import { Breadcrumb, Layout, Menu, theme } from 'antd';
import { MdOutlineSatelliteAlt } from "react-icons/md";
import { CreateNamespaceForm } from '../Component/CreateNamespaceForm';
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

var consoleItems = [
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

const siderMemuMap = {
  "console": consoleItems
  
}
const App = () => {
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  var [selectedTopBarItem,setSelectedTopBarItem] = React.useState("console");
  var [selectedSiderItem,setSelectedSiderItem] = React.useState("add_namespace");
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
            } else {
              selectedSiderItem = ""
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
            <CreateNamespaceForm/>
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};
export default App;