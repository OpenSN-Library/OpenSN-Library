import {Card,List,Button,Typography} from "antd";

export function ConfigNamespaceLink({dataBuf,setDataBuf}) {
    return (
        <Card title={"命名空间初始链路配置"}>
            <List
                bordered
                header={<div style={{height:"30px"}}>
                        <Typography.Text strong style={{float:"left"}}>命名空间链路列表</Typography.Text>
                        <Button style={{float:"right"}}>添加</Button>
                        </div>}
                dataSource={dataBuf.link_config}
                renderItem={(item,index) => 
                    <List.Item>
                    <List.Item.Meta
                        title={<Typography.Text>序号:{index}-类型:{item.type}</Typography.Text>}
                        description={<div>
                            <div>
                                <Typography.Text>
                                    {`连接信息:${item.instance_index[0]}<-->${item.instance_index[1]}`}
                                </Typography.Text>
                            </div>
                            <Typography.Text>参数</Typography.Text>
                            <div>
                                {JSON.stringify(item.parameter)}
                            </div>
                        </div>}
                    />
                    <div>
                        <Button>编辑</Button> 
                        <Button>删除</Button>
                    </div>
                    </List.Item>
                }
            />
        </Card>
    )
}