import { useEffect,useState } from "react";
import { GetDatabaseItems } from "../Request/databse";
import { Button, Card, Col, Divider, Row, Typography,Tree } from "antd";

import { DeleteDatabaseItem,UpdateDatabaseItem } from "../Request/databse";
import TextArea from "antd/es/input/TextArea";


export const FilePage = () => {
  const [showLine, setShowLine] = useState(true);
  const [showIcon, setShowIcon] = useState(false);
  const [showLeafIcon, setShowLeafIcon] = useState(true);
  const [databaseItems,setDatabaseItems] = useState({})
  const [treeKeys,setTreeKeys] = useState([])
  const [selectKey,setSelectKey] = useState("/")
  const [editorContent,setEditorContent] = useState({
      json:{},
      text:""
  })
  useEffect(()=>{
      GetDatabaseItems((response)=>{
          const data = response.data.data
          const tmpTreeKeys = {}
          const items = {}
          Object.keys(data).forEach((item)=>{
              const keys = item.split("/")
              let tmp = tmpTreeKeys
              keys.forEach((key)=>{
                  if(key===""){
                      return
                  }
                  if(tmp[key]===undefined){
                      tmp[key] = {
                          realKey : keys.slice(0,keys.indexOf(key)+1).join("/"),
                          children:{}
                      }
                  }
                  tmp = tmp[key].children
              })
              items[item] = data[item]
          })
          console.log(tmpTreeKeys)
          setTreeKeys(tmpTreeKeys)
          setDatabaseItems(items)      })
  },[])

  const transTreeData = (data) => {
      const res = []
      Object.keys(data).forEach((key)=>{
          const item = data[key]
          res.push({
              key:data[key].realKey,
              title:key,
              children:transTreeData(item.children)
          })
      })
      return res
  }

  const onSelect = (selectedKeys, info) => {
      console.log('selected', selectedKeys, info);
      setSelectKey(selectedKeys[0])
      
  }


  return (
      <Card style={{height:"85vh"}}>
          <Row justify="space-between" style={{height:"100%"}} >
              <Col style={{height:"100%"}}>
                  <Card style={{width:"20vw",height:"80vh"}}>
                      <Typography.Title level={4}>数据库键列表</Typography.Title>
                      <Tree
                          style={{height:"65vh",overflowY:"auto"}}
                          showLine={showLine ? { showLeafIcon } : false}
                          showIcon={showIcon}
                          defaultExpandedKeys={['/']}
                          onSelect={onSelect}
                          treeData={transTreeData(treeKeys)}
                      />
                  </Card>
              </Col>
              <Col style={{height:"100%"}}>
              
              <Card style={{width:"70vw",height:"80vh"}}>
              <Typography.Title level={4}>文件预览(仅文本文件)</Typography.Title>
              <Row justify="right"></Row>
                  <Card style={{maxHeight:"60vh",overflowY:"scroll"}}>
                  <Typography.Text style={{height:"80%",height:"80%"}} >
                    
                  </Typography.Text>

                  </Card>
                  
                  <Divider/>
                  <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                      onClick={()=>{
                          UpdateDatabaseItem({
                              key:selectKey,
                              value:editorContent.text
                          },(response)=>{
                              console.log(response)
                          })
                      }}
                  >下载</Button>
              </Card>
              </Col>
          </Row>
      </Card>
  );
}