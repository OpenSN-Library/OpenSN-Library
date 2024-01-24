import {Card,List,Button,Input,Typography,Divider} from "antd";
import { ContainerEnvDrawer } from "./ContainerEnvDrawer";
import { useState } from "react";
import { InstanceTypeDrawer } from "./InstanceTypeDrawer";
export function ConfigNamespaceBasic({dataBuf,setDataBuf}) {
    const [envMapOpen,setEnvMapOpen] = useState(false);
    const [instanceTypeOpen,setInstanceTypeOpen] = useState(false);
    return (

            <Card title={"命名空间基本信息"}>
                <Typography.Text strong>名称: </Typography.Text>
                <Input type="text" style={{width:"200px"}} onChange={(e) => {
                    dataBuf.name = e.target.value;
                    setDataBuf(dataBuf);
                }} />
                <Divider dashed />
                
                <List
                    size="small"
                    bordered
                    header={<div style={{height:"30px"}}>
                        <Typography.Text strong style={{float:"left"}}>实例类型与镜像</Typography.Text>
                        <Button 
                            style={{float:"right"}}
                            onClick={()=>{setInstanceTypeOpen(true)}}
                        >
                            添加
                        </Button>
                        <InstanceTypeDrawer
                            dataBuf={dataBuf} setDataBuf={setDataBuf}
                            open={instanceTypeOpen} setOpen={setInstanceTypeOpen}
                        />
                        </div>}
                    dataSource={Object.keys(dataBuf.ns_config.image_map)}
                    style={{height:"320px",overflow:"auto"}}
                    renderItem={(item_key) => (
                    <List.Item key={item_key}>
                        <Card title={`实例类型:${item_key}`}>
                            <Typography.Paragraph>
                                {`镜像名称:${dataBuf.ns_config.image_map[item_key]}`}
                            </Typography.Paragraph>
                            <Typography.Paragraph>
                                {`CPU配额(1e-9):${dataBuf.ns_config.resource_map[item_key].nano_cpu}`}
                            </Typography.Paragraph>
                            <Typography.Paragraph> 
                                {`内存配额(Byte):${dataBuf.ns_config.resource_map[item_key].memory_byte}`}
                            </Typography.Paragraph>
                            <Button 
                                style={{float:"right"}}
                                onClick={()=>{
                                    delete dataBuf.ns_config.image_map[item_key]
                                    delete dataBuf.ns_config.resource_map[item_key]
                                    // setDataBuf(dataBuf)
                                }}
                            >
                                删除
                            </Button>
                            
                        </Card>
                    </List.Item>
                    )}
                />
                <Divider dashed />
                <List
                    size="small"
                    bordered
                    style={{height:"240px",overflow:"auto"}}
                    header={<div style={{height:"30px"}} >
                        <Typography.Text strong style={{float:"left"}}>初始化环境变量</Typography.Text>
                        <Button 
                            style={{float:"right"}}
                            onClick={()=>{setEnvMapOpen(true)}}
                        >
                            添加
                        </Button>
                        <ContainerEnvDrawer 
                            dataBuf={dataBuf} setDataBuf={setDataBuf}
                            open={envMapOpen} setOpen={setEnvMapOpen}
                        />
                    </div>}
                    dataSource={Object.keys(dataBuf.ns_config.container_envs)}
                    renderItem={(item) => <List.Item>
                        {`${item}=${dataBuf.ns_config.container_envs[item]}`} 
                        <Button style={{float:"right"}} >删除</Button>
                        </List.Item>}
                />
                </Card>
        
    )
}