import { Button, Card, Divider, List,Typography } from "antd";
import { useEffect, useState } from "react";
import { GetNodeList } from "../Request/node";
import { GetInstanceList } from "../Request/instance";
import { Link } from "react-router-dom";
export const InstanceListPage = () => {
    let [instanceList,setInstanceList] = useState([])
    useEffect(()=>{
        GetInstanceList((response)=>{
        setInstanceList(response.data.data==null?[]:response.data.data)
    })
    },[])
    return (
        <List
            itemLayout="vertical"
            size="large"
            pagination={{
            onChange: (page) => {
                console.log(page);
            },
            pageSize: 16,
            }}
            header = {
                <div style={{height:"70px"}} >
                    <Typography.Title level={4}>网络节点列表</Typography.Title>
                    <Button>添加节点</Button>
                </div>
                
            }
            dataSource={instanceList}
            renderItem={(item,index)=>{
                return (
                    <List.Item key={item.link_id} actions={[
                        
                        <Button
                            onClick={()=>{
                                const w = window.open('_black')
                                let url = `/instance/${item.node_index}/${item.instance_id}`
                                w.location.href = url
                            }}
                        >查看详情</Button>,

                        <Button
                            onClick={item.enable?
                                ()=>{}:
                                ()=>{}
                            }
                        >{item.start?"停止":"启动"}</Button>,,
                        <Button>删除</Button>
                    ]}>
                        <List.Item.Meta title={`机器编号${item.node_index} ${item.type}-${item.instance_id}: ${item.start?"已启动":"已停止"}`}/>
                        
                    </List.Item>
                )
            }}
        />
    )
}