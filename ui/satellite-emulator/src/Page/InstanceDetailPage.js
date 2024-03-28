
import { Card, Col, Descriptions, Divider, List, Row, Typography, Badge, Button,Tabs} from "antd"
import { useEffect, useState } from "react"
import { useParams } from "react-router-dom"
import ReactECharts from 'echarts-for-react'
import { GetPeriodInstanceResource } from "../Request/metrics"
import { GetInstanceDetail } from "../Request/instance"

export const InstanceDetailPage = () => {
    const instanceID = useParams().instance_id
    const nodeIndex = useParams().node_index
    const [instanceResource,setInstanceResource] = useState({})
    const [resourcePeriod,setResourcePeriod] = useState("10m")
    const [instanceInfo,setInstanceInfo] = useState({})
    const [items, setItems] = useState([
    ]);
    const [activeKey, setActiveKey] = useState("");
    useEffect(()=>{
        GetPeriodInstanceResource(instanceID,resourcePeriod,(response)=>{
            setInstanceResource(response.data.data)
        })
        GetInstanceDetail(nodeIndex,instanceID,(response)=>{
            setInstanceInfo(response.data.data)
        })
    },[])
    
    const descriptionDetail = [
        {
            key:"instance_id",
            label:"节点ID",
            children:instanceID,
        },
        {
            key:"instance_name",
            label:"节点名称",
            children:instanceInfo.name,
        },
        {
            key:"instance_start",
            label:"节点启动状态",
            children:instanceInfo.start?
                <Badge status="success" text="已启动" />:
                <Badge status="default" text="未启动" />    
            ,
        },
        {
            key:"instance_type",
            label:"节点类型",
            children:instanceInfo.type,
        },
        {
            key:"instance_image",
            label:"节点镜像",
            children:instanceInfo.image,
        },
        {
            key:"node_index",
            label:"节点部署机器编号",
            children:nodeIndex,
        },
        {
            key:"instance_resource",
            label:"节点资源限制",
            children:<div>
                <p>
                    CPU限额(Nano): {instanceInfo?.resource_limit?.nano_cpu}
                </p>
                <p>
                    内存(Byte): {instanceInfo?.resource_limit?.memory_byte}
                </p>
            </div>,
        },
        {
            key:"instance_connections",
            label:"节点连接信息",
            children:<List>
                {instanceInfo.connections?Object.keys(instanceInfo.connections).map((item,index)=>{
                    return (
                        <List.Item key={index}>
                            <List.Item.Meta title={<Typography.Link href={`/link/${instanceInfo.node_index}/${item}`} target="_blank">{item}</Typography.Link>}/>
                            <p>
                                {`连接到部署于机器${instanceInfo.connections[item].end_node_index}的类型为${instanceInfo.connections[item].instance_type}节点`}
                                <Typography.Link  href={`/instance/${instanceInfo.connections[item].end_node_index}/${instanceInfo.connections[item].instance_id}`} target="_blank">{instanceInfo.connections[item].instance_id} </Typography.Link>
                            </p>
                        </List.Item>
                    )
                }):[]}
            </List>,
        },
    ]

    const cpuChartOption = {
        title: {
            text: '节点CPU使用情况'
        },
        tooltip: {},
        legend: {
            data:['CPU使用率']
        },
        xAxis: {
            data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.Time):[]
        },
        yAxis: {},
        series: [{
            name: 'CPU使用率',
            type:'line',
            data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.cpu_usage):[]
        }]
    }

    const memChartOption = {
        title: {
            text: '节点内存使用情况'
        },
        tooltip: {},
        legend: {
            data:['内存使用量','交换内存使用量']
        },
        xAxis: {
            data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.Time):[]
        },
        yAxis: {},
        series: [
            {
                name: '内存使用量',
                type:'line',
                data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.mem_byte):[]
            },
            {
                name: '交换内存使用量',
                type:'line',
                data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.swap_mem_byte):[]
            }
        ]
        
    }

    return (
        <div>
            <Row justify="center">
                <Col>
                    <Typography.Title level={2} >实例详情:{instanceID}</Typography.Title>
                </Col>
            </Row>
            <Row justify="center">
                <Card>
                    <Col>
                    <Descriptions bordered items={descriptionDetail} />
                    </Col>
                </Card>
                
            </Row>
            <Divider/>
            <Row justify="space-between">
                
                    <Col style={{marginLeft:"5vw"}}>
                        <Card>
                            <ReactECharts
                                option={cpuChartOption}
                                style={{ height: "30vh",width:"40vw" }}
                            />
                        </Card>
                    </Col>
                    <Col style={{marginRight:"5vw"}}>
                        <Card>
                            <ReactECharts
                                option={memChartOption}
                                style={{ height: "30vh",width:"40vw" }}
                            />
                        </Card>
                    </Col>
                
            </Row>
            <Divider/>
            
            <Row justify="center" >
            <Typography.Title level={4}>WebShell</Typography.Title>
            <Card style={{width:"90vw",marginLeft:"5vw",marginRight:"5vw"}}>
                <Tabs
                    type="editable-card"
                    onChange={(newActiveKey) => {
                        setActiveKey(newActiveKey);
                    }}
                    style={{width:"100%"}}
                    activeKey={activeKey}
                    onEdit={(targetKey, action)=>{
                        if (action === 'add') {
                            console.log("add")
                            const newKey = `${instanceID}-${items.length}`;
                            setItems([...items, { 
                                key: newKey,
                                label: newKey,
                                children: <Row justify="center" ><iframe style={{width:"80vw",height:"720px"}} src="http://10.134.148.56:8079"/></Row>
                            }]);
                            setActiveKey(newKey);
                        } else {
                            let newActiveKey = activeKey;
                            let lastIndex;
                            items.forEach((item, i) => {
                                if (item.title === targetKey) {
                                    lastIndex = i - 1;
                                }
                            });
                            const newPanes = items.filter(item => item.title !== targetKey);
                            if (newPanes.length && newActiveKey === targetKey) {
                                if (lastIndex >= 0) {
                                    newActiveKey = newPanes[lastIndex].title;
                                } else {
                                    newActiveKey = newPanes[0].title;
                                }
                            }
                            setItems(newPanes);
                            setActiveKey(newActiveKey);
                        }
                    }}
                    items={items}
                />
                </Card>
            </Row>
        </div>
    )
}