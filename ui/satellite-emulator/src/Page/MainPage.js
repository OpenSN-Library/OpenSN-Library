import {React,useState} from 'react';
import {Layout, Menu, theme } from 'antd';
import { MdOutlineSatelliteAlt } from "react-icons/md";
import { CreateNamespaceForm } from '../Component/CreateNamespaceForm';
import { NamespaceList } from '../Component/GetNamespaceList';
import { NodeList } from '../Component/GetNodeList';
import { ConsoleItems } from './ConsolePage';
import { FileItems } from './FilePage';
import ExportOutlined from '@ant-design/icons/ExportOutlined';
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






const componentMap = {
    "add_namespace": <CreateNamespaceForm/>,
    "namespace_list": <NamespaceList/>,
    "node_list": <NodeList/>,
}

const siderMemuMap = {
  "console": ConsoleItems,
  "mount_manager": FileItems,
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
            if (key in siderMemuMap) {
              setSelectedTopBarItem(key);
              if (siderMemuMap[key] != null && siderMemuMap[key].length > 0){
                selectedSiderItem = siderMemuMap[key][0].key
                setSelectedSiderItem(selectedSiderItem);
              } else {
                selectedSiderItem = ""
                setSelectedSiderItem(selectedSiderItem);
              }
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