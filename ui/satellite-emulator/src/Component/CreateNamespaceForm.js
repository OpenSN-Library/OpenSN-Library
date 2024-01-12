
import { Form ,Button, Input,List,Card,} from "antd";
import React, { useState } from "react";
import { CreateNamespaceReq } from "../Model/Namespace";


export function CreateNamespaceForm() {
    var [dataBuf,setDataBuf] = useState(new CreateNamespaceReq());
    return (
        <div>
            <Form label="创建命名空间">
                <Form.Item label="命名空间名称">
                    <Input type="text" onChange={(e) => {
                        dataBuf.name = e.target.value;
                        setDataBuf(dataBuf);
                    }} />
                </Form.Item>
                <Form.Item label="实例类型与对应镜像">
                    <Button>添加</Button>
                    <List
                        dataSource={dataBuf.ns_config.image_map.keys}
                        renderItem={(item_key) => (
                        <List.Item>
                            <Card title={item_key}>
                                {dataBuf.ns_config.image_map[item_key]}
                                <Button>删除</Button>
                            </Card>
                        </List.Item>
                        )}
                    />
                </Form.Item>
                <Form.Item label="预设环境变量">
                <Button>添加</Button>
                </Form.Item>
            </Form>
        </div>
    )
}