import { Button, Card, Divider, List,Typography } from "antd";
import { useEffect, useState } from "react";
import { GetNodeList } from "../Request/node";
import { GetLinkList } from "../Request/link";
export const LinkListPage = () => {
    const [linkList,setLinkList] = useState([])
    const [parameter,setParameter] = useState({})
    useEffect(()=>{
        GetLinkList((response)=>{
            setLinkList(response.data.data==null?[]:response.data.data)
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
                    <Typography.Title level={4}>链路列表</Typography.Title>
                    <Button>添加链路</Button>
                </div>
                
            }
            dataSource={linkList}
            renderItem={(item,index)=>{
                return (
                    <List.Item key={item.link_id} actions={[
                        <Button
                            onClick={()=>{
                                const w = window.open('_black')
                                let url = `/link/${item.link_id}`
                                w.location.href = url
                            }}
                        >查看详情</Button>,
                        <Button
                            onClick={item.enable?
                                ()=>{}:
                                ()=>{}
                            }
                        >{item.enable?"禁用":"启用"}</Button>,
                        <Button>删除</Button>
                    ]}>
                        <List.Item.Meta title={`${item.type}-${item.link_id}: ${item.enable?"已启用":"已禁用"}`}/>
                        
                    </List.Item>
                )
            }}
        />
    )
    
}