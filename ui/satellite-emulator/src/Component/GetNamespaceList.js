import { Button, Card, Divider, List,Typography } from "antd";
import { useEffect, useState } from "react";
import { GetNamespaceList,StartNamespace,StopNamespace } from "../Request/Namespace";
export const NamespaceList = () => {
    let [namespaceList,setNamespaceList] = useState([])
    useEffect(()=>{
        GetNamespaceList((response)=>{
            setNamespaceList(response.data.data==null?[]:response.data.data)
        })
    },[])
    return (
        <List
                bordered
                header={<div style={{height:"30px"}}>
                        <Typography.Text strong style={{float:"left"}}>命名空间列表</Typography.Text>
                        </div>}
                direction="vertical"
                dataSource={namespaceList}
                renderItem={(item,index) => 
                    <List.Item>
                    <Card>
                        <Card.Meta
                            title={`命名空间名称: ${item.name}-${item.running?"运行中":"停止中"}`}
                        />
                        <Divider/>
                        <div>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`命名空间容器实例数:${item.instance_num}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`命名空间链路数:${item.link_num}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`命名空间分配计算节点列表:${item.running?JSON.stringify(item.alloc_node_index):"未运行"}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                       
                        </div>
                        
                        <div>
                            <Button
                                onClick={()=>{
                                    if (item.running) {
                                        StopNamespace(item.name,(response)=>{
                                            console.log(response)
                                            GetNamespaceList((response)=>{
                                                setNamespaceList(response.data.data)
                                            })
                                        })
                                    } else {
                                        StartNamespace(item.name,(response)=>{
                                            console.log(response)
                                            GetNamespaceList((response)=>{
                                                setNamespaceList(response.data.data)
                                            })
                                        })
                        
                                    }
                                }}
                            >
                                {item.running?"停止":"启动"}
                            </Button>
                            <Button>编辑</Button>
                            <Button
                                onClick={()=>{
                                    window.open(`/namespace/${item.name}/detail`, "_blank")
                                }}
                            >详细信息</Button>
                        </div>
                    </Card>
                   
                    </List.Item>}
            />
    )
}