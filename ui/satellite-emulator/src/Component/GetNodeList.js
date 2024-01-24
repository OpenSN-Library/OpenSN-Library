import { Button, Card, Divider, List,Typography } from "antd";
import { useEffect, useState } from "react";
import { GetNodeList } from "../Request/node";
export const NodeList = () => {
    let [nodeList,setNodeList] = useState([])
    useEffect(()=>{
        GetNodeList((response)=>{
            setNodeList(response.data.data==null?[]:response.data.data)
        })
    },[])
    return (
        <List
                bordered
                header={<div style={{height:"30px"}}>
                        <Typography.Text strong style={{float:"left"}}>计算节点列表</Typography.Text>
                        </div>}
                dataSource={nodeList}
                renderItem={(item,index) => 
                    <List.Item>
                    <Card>
                        <Card.Meta
                            title={`节点编号: ${item.node_id}-${item.is_master_node?"主节点":"从节点"}`}
                        />
                        <Divider/>
                        <div>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`剩余可分配容器数:${item.free_instance}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`IPv4地址:${item.l_3_addr_v_4}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`IPv6地址:${item.l_3_addr_v_6}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        <Typography.Paragraph>
                            <Typography.Text>
                                {`MAC地址:${item.l_2_addr}`}
                            </Typography.Text>
                        </Typography.Paragraph>
                        </div>
                        
                        <div>
                            <Button>详细信息</Button>
                        </div>
                    </Card>
                   
                    </List.Item>}
            />
    )
}