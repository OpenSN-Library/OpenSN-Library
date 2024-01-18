import {Card,List,Button,Input,Typography,Divider} from "antd";
export function ConfigNamespaceInstance({dataBuf,setDataBuf}) {
    return (
        <Card title={"命名空间初始节点配置"}>
            <List
                bordered
                header={<div style={{height:"30px"}}>
                        <Typography.Text strong style={{float:"left"}}>命名空间容器实例列表</Typography.Text>
                        <Button style={{float:"right"}}>添加</Button>
                        </div>}
                dataSource={dataBuf.inst_config}
                renderItem={(item,index) => 
                    <List.Item>
                    <List.Item.Meta
                        title={<Typography.Text>序号:{index}-类型:{item.type}</Typography.Text>}
                        description={<div>
                            <Typography.Text>附加信息</Typography.Text>
                            <div>
                                {JSON.stringify(item.extra)}
                            </div>
                        </div>}
                    />
                    <div>
                        <Button>编辑</Button> 
                        <Button>删除</Button>
                    </div>
                    </List.Item>}
            />
        </Card>
    )
}