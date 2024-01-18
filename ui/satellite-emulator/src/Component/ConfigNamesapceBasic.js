import {Card,List,Button,Input,Typography,Divider} from "antd";

export function ConfigNamespaceBasic({dataBuf,setDataBuf}) {
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
                        <Button style={{float:"right"}}>添加</Button>
                        </div>}
                    dataSource={Object.keys(dataBuf.ns_config.image_map)}
                    style={{height:"240px",overflow:"auto"}}
                    renderItem={(item_key) => (
                    <List.Item>
                        <Card title={item_key}>
                            {dataBuf.ns_config.image_map[item_key]}
                            <Button style={{float:"right"}} >删除</Button>
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
                        <Button style={{float:"right"}}>添加</Button>
                    </div>}
                    dataSource={Object.keys(dataBuf.ns_config.container_envs)}
                    renderItem={(item) => <List.Item>
                        {item=dataBuf.ns_config.container_envs[item]} 
                        <Button style={{float:"right"}} >删除</Button>
                        </List.Item>}
                />
                </Card>
        
    )
}