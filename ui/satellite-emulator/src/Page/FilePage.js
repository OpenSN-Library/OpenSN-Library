import { useEffect,useState,useRef } from "react";
import { GetDatabaseItems } from "../Request/databse";
import { Button, Card, Col, Divider, Row, Typography,Tree } from "antd";

import { DeleteDatabaseItem,UpdateDatabaseItem } from "../Request/databse";
import TextArea from "antd/es/input/TextArea";
import { UrlBase } from "../Request/base";
import { DeleteFile, DownloadFile, GetFileList, PreivewFile, UploadFile } from "../Request/file";
import { message } from "antd";

export const FilePage = () => {
  const inputRef = useRef(null)
  const [messageApi, contextHolder] = message.useMessage();
  const [treeKeys,setTreeKeys] = useState([])
  const [selectPath,setSelectPath] = useState("")
  const [previewContent,setPreviewContent] = useState("")
  const [loadedKeys,setLoadedKeys] = useState([])
  useEffect(()=>{
      GetFileList(selectPath, (response)=>{
          const data = response.data.data?response.data.data:[]
          const tmpTreeKeys = [{
            title:"",
            key:"",
            children:[],
            isLeaf:false
        }]

          setTreeKeys(tmpTreeKeys)
        })
  },[])

  const onSelect = (selectedKeys, info) => {
      if (selectedKeys.length!==1){
          return
      }
      setSelectPath(selectedKeys[0])
      PreivewFile(selectedKeys[0],(response)=>{
            setPreviewContent(response.data.data)
      })
  }


  return (
      <Card style={{height:"85vh"}}>
          <Row justify="space-between" style={{height:"100%"}} >
              <Col style={{height:"100%"}}>
                  <Card style={{width:"20vw",height:"80vh"}}>
                      <Typography.Title level={4}>文件列表</Typography.Title>
                      <input type="file" ref={inputRef} style={{display:"none"}} onChange={(e)=>{
                          const file = e.target.files[0]
                          UploadFile(file,selectPath,(response)=>{
                            messageApi.open({
                                type: "success",
                                content: "上传文件成功",
                            })
                          })
                      } }/>
                      <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                          onClick={()=>{
                            inputRef.current.click()
                          }}
                      >上传</Button>
                      <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}} 
                        onClick={()=>{
                          setSelectPath("")
                          const tmpTreeKeys = [{
                            title:"",
                            key:"",
                            children:[],
                            isLeaf:false,
                            loaded:false  
                          }]
                          setLoadedKeys([])
                          setTreeKeys(tmpTreeKeys)
                          
                        }}
                      >刷新</Button>
                      <Tree
                          loadedKeys={loadedKeys}
                          style={{height:"65vh",overflowY:"auto"}}
                          loadData={({key,children})=>{
                            return new Promise((resolve) => {
                                GetFileList(key, (response)=>{
                                    const data = response.data.data?response.data.data:[]
                                    const tmpTreeKeys = data.map((item)=>{
                                          return({
                                            title:item.name,
                                            key:`${key}${item.name}${item.is_dir?"/":""}`,
                                            children:[],
                                            isLeaf:!item.is_dir
                                          })
                                    })
                                    children.length = 0
                                    children.push(...tmpTreeKeys)
                                  
                                    console.log(treeKeys)
                                    setTreeKeys([...treeKeys])
                                    resolve();
                                })
                              })
                          }}
                          onSelect={onSelect}
                          treeData={treeKeys}
                      />
                  </Card>
              </Col>
              <Col style={{height:"100%"}}>
              
              <Card style={{width:"70vw",height:"80vh"}}>
              <Typography.Title level={4}>文件预览(仅文本文件)</Typography.Title>
              <Row justify="right"></Row>
                  <TextArea wrap="false" readOnly style={{height:"60vh"}} value={previewContent}/>

                  
                  <Divider/>
                  <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                      onClick={()=>{
                          
                          const downloadUrl = UrlBase + `/file/download/${selectPath.split("/").pop()}?path=${selectPath}`
                          window.open(downloadUrl)
                          messageApi.open({
                              type: "success",
                              content: "下载文件成功",
                          })
                      }}
                  >下载</Button>
                  <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                      onClick={()=>{
                          DeleteFile(selectPath,(response)=>{
                            messageApi.open({
                                type: "success",
                                content: "删除文件成功",
                            })
                          })
                      }}
                  >删除</Button>
              </Card>
              </Col>
          </Row>
      </Card>
  );
}