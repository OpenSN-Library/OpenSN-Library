
import { Tabs, Typography, Divider, Button, notification} from "antd";
import React, { useState, useRef } from "react";
import { ConfigNamespaceBasic } from "./ConfigNamesapceBasic";
import { ConfigNamespaceInstance } from "./ConfigNamespaceInstance";
import { ConfigNamespaceLink } from "./ConfigNamespaceLink";

export function CreateNamespaceForm() {
    const [dataBuf,setDataBuf] = useState({});

    const [api, contextHolder] = notification.useNotification();
    const NotificationWithIcon = (type,msg,detail) => {
        api[type]({
            message: msg,
            description:detail,
        });
    };
    const inputRef = useRef(null);
    return (
        <div style={{height:"100%",overflowY:"scroll"}}>
            {contextHolder}
            <Typography.Title level={2}>创建命名空间</Typography.Title>
            <Divider dashed />
            <Tabs defaultActiveKey="1" 
                tabPosition="left" 
                tabBarExtraContent={{
                        "left" : 
                            <div>
                                <div>
                                    <input
                                        style={{display: 'none'}}
                                        ref={inputRef}
                                        type="file"
                                        onChange={(e) => {
                                            try {
                                                const fileObj = e.target.files && e.target.files[0];
                                                console.log(fileObj);
                                                const reader = new FileReader();
                                                reader.onload = function (e) {
                                                    try {
                                                        dataBuf = JSON.parse(e.target.result)
                                                        setDataBuf(dataBuf);
                                                    }catch (e) {
                                                        console.error(e);
                                                        NotificationWithIcon("error","导入失败","Error:"+e.toString());
                                                    }
                                                };
                                                reader.readAsText(fileObj);
                                                NotificationWithIcon("success","导入成功","成功导入配置文件");
                                            } catch (e) {
                                                console.error(e);
                                                NotificationWithIcon("error","导入失败","Error:"+e.toString());
                                            }
                                        }}
                                    />
                                    <Button 
                                        style={{marginBottom:"15px"}}
                                        onClick={() => {
                                            inputRef.current.click();
                                        }}
                                    >
                                        导入
                                    </Button>
                                </div>
                                <div>
                                    <Button 
                                        style={{marginBottom:"15px"}}
                                        onClick={() => {
                                            const blob = new Blob([JSON.stringify(dataBuf)], {
                                                type: 'application/json'
                                            })
                                            const objectURL = URL.createObjectURL(blob)
                                            const domElement = document.createElement('a')
                                            domElement.href = objectURL
                                            domElement.download = "namespace_"+dataBuf.name+"_" + Date.now() + ".json"
                                            domElement.click()
                                            URL.revokeObjectURL(objectURL)
                                        }}
                                    >
                                        导出
                                    </Button>
                                </div>
                            </div>,
                    "right":<Button onClick={()=>{
                        
                            NotificationWithIcon("success","创建成功","成功创建命名空间")
                        
                    }}> 
                        确认
                    </Button>  
                   }}
                items = {[
                    {
                        key :"Basic",
                        label :"基本配置",
                        children : <div >
                            <ConfigNamespaceBasic  dataBuf={dataBuf} setDataBuf={setDataBuf}/>
                        </div> 
                    },
                    {
                        key :"Instance",
                        label :"容器实例配置",
                        children : <div >
                            <ConfigNamespaceInstance  dataBuf={dataBuf} setDataBuf={setDataBuf}/>
                        </div> 
                    },
                    {
                        key :"Link",
                        label :"链路实例配置",
                        children : <div >
                            <ConfigNamespaceLink  dataBuf={dataBuf} setDataBuf={setDataBuf}/>
                        </div> 
                    }
                ]
            }/>
        </div>     
    )
}