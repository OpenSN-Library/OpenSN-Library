
import { Card, Col, Descriptions, Divider, List, Row, Typography, Badge, Button,Select} from "antd"
import { useEffect, useState } from "react"
import { useParams } from "react-router-dom"
import ReactECharts from 'echarts-for-react'
import { GetPeriodInstanceResource, GetPeriodLinkResource } from "../Request/metrics"
import { GetInstanceDetail } from "../Request/instance"
import { StartLinkWebshell } from "../Request/webshell"
import { GetLinkInfo, GetLinkParameter } from "../Request/link"

export const LinkDetailPage = () => {
    const linkID = useParams().link_id
    const nodeIndex = useParams().node_index
    const [linkResource,setLinkResource] = useState({})
    const [linkParameter,setLinkParameter] = useState({})
    const [resourcePeriod,setResourcePeriod] = useState("10m")
    const [linkInfo,setLinkInfo] = useState({})
    const [linkDumpUrl,setLinkDumpUrl] = useState("")
    useEffect(()=>{
        GetLinkInfo(nodeIndex,linkID,(response)=>{
            setLinkInfo(response.data.data)
        })
        GetPeriodLinkResource(linkID,resourcePeriod,(response)=>{
            setLinkResource(response.data.data)
        })
        GetLinkParameter(nodeIndex,linkID,(response)=>{
            setLinkParameter(response.data.data)
        })
        setInterval(()=>{
            GetLinkInfo(nodeIndex,linkID,(response)=>{
                setLinkInfo(response.data.data)
            })
            GetPeriodLinkResource(linkID,resourcePeriod,(response)=>{
                setLinkResource(response.data.data)
            })
            GetLinkParameter(nodeIndex,linkID,(response)=>{
                setLinkParameter(response.data.data)
            })
        },1000)
    },[])
    
    const descriptionDetail = [
        {
            key:"link_id",
            label:"链路ID",
            children:linkID,
        },
        {
            key:"link_enable",
            label:"链路启用状态",
            children:linkInfo.enable?
                <Badge status="success" text="已启动" />:
                <Badge status="default" text="未启动" />    
            ,
        },
        {
            key:"link_type",
            label:"链路类型",
            children:linkInfo.type,
        },
        {
            key:"node_index",
            label:"链路部署机器编号",
            children:nodeIndex,
        },
        {
            key:"link_parameter",
            label:"链路参数",
            children:<List>
                {Object.keys(linkParameter).map((key)=>{
                    return (
                        <List.Item key={key}>
                            <List.Item.Meta
                                title={key}
                                description={linkParameter[key]}
                            />
                        </List.Item>
                    )
                })}
            </List>,
        },
        {
            key:"link_end_infos",
            label:"链路两端信息",
            children:<List>
                {linkInfo.end_infos?linkInfo.end_infos.map((item,index)=>{
                    return (
                        <List.Item key={index}>
                            <List.Item.Meta
                                title={<Typography.Link href={`/instance/${item.end_node_index}/${item.instance_id}`} target="_blank">{item.instance_id}</Typography.Link>}
                                description={`部署于机器${item.node_index}的类型为${item.instance_type}节点`}
                            />
                        </List.Item>
                    )
                }):[]}
            </List>,
        },
        {
            key:"link_addr_infos",
            label:"链路端地址信息",
            children:<List bordered>
                {linkInfo.address_infos?linkInfo.address_infos.map((item,index)=>{
                    return (
                        <List.Item key={index}>
                            {Object.keys(item).map((key)=>{
                                return (
                                    <p key={key}>
                                        {`${key}= ${item[key]}`}
                                    </p>
                                )
                            })}
                        </List.Item>
                    )
                }):[]}
            </List>,
        }
    ]

    const cpuChartOption = {
        title: {
            text: '链路收发吞吐量(Bps)'
        },
        tooltip: {},
        legend: {
            data:['接收数据','发送数据']
        },
        xAxis: {
            data: linkResource[linkID]?linkResource[linkID].map((item)=>{
                const time = Date.parse(item.Time)
                return `${new Date(time).getHours()}:${new Date(time).getMinutes()}:${new Date(time).getSeconds()}`
            }):[]
        },
        yAxis: {},
        series: [
            {
                name: '发送数据量',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.send_bps):[]
            },
            {
                name: '接收数据量',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.recv_bps):[]
            }
        ]
    }

    const memChartOption = {
        title: {
            text: '链路数据包统计'
        },
        tooltip: {},
        legend: {
            data:['发送数据包','接收数据包','发送丢包量','接收丢包量','发送错误包','接收错误包']
        },
        xAxis: {
            data: linkResource[linkID]?linkResource[linkID].map((item)=>{
                const time = Date.parse(item.Time)
                return `${new Date(time).getHours()}:${new Date(time).getMinutes()}:${new Date(time).getSeconds()}`
            }):[]
        },
        yAxis: {},
        series: [
            {
                name: '发送数据包',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.send_pps):[]
            },
            {
                name: '接收数据包',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.recv_pps):[]
            },
            {
                name: '发送丢包量',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.send_drop_pps):[]
            },
            {
                name: '接收丢包量',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.recv_drop_pps):[]
            },
            {
                name: '发送错误包',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.send_err_pps):[]
            },
            {
                name: '接收错误包',
                type:'line',
                data: linkResource[linkID]?linkResource[linkID].map((item)=>item.recv_err_pps):[]
            }
        ]
        
    }

    return (
        <div>
            <Row justify="center">
                <Col>
                    <Typography.Title level={2} >链路详情:{linkID}</Typography.Title>
                </Col>
            </Row>
            <Row justify="center">
                <Card>
                    <Col>
                    <Descriptions style={{width:"90vw"}} bordered items={descriptionDetail} />
                    </Col>
                </Card>
                
            </Row>
            <Divider/>
            <Row justify="center"> <Typography.Title level={4}>资源监控</Typography.Title> </Row>
            <Row justify="center"> 
            <Select
                defaultValue={resourcePeriod}
                style={{ width: 120 }}
                onChange={(value)=>{
                    setResourcePeriod(value)
                    GetPeriodLinkResource(linkID,resourcePeriod,(response)=>{
                        setLinkResource(response.data.data)
                    })
                }}
                options={[
                    { value: '1m', label: '过去一分钟' },
                    { value: '5m', label: '过去五分钟' },
                    { value: '10m', label: '过去十分钟' },
                    { value: '30m', label: '过去三十分钟' },
                    { value: '1h', label: '过去一小时' },
                    { value: '3h', label: '过去三小时' },
                    { value: '6h', label: '过去六小时' },
                    { value: '12h', label: '过去十二小时' },
                    { value: '24h', label: '过去二十四小时' },
                ]}
            />
            </Row>
            <Row justify="space-between">
                    <Col style={{marginLeft:"5vw"}}>
                        <Card>
                            <ReactECharts
                                option={cpuChartOption}
                                style={{ height: "40vh",width:"40vw" }}
                            />
                        </Card>
                    </Col>
                    <Col style={{marginRight:"5vw"}}>
                        <Card>
                            <ReactECharts
                                option={memChartOption}
                                style={{ height: "40vh",width:"40vw" }}
                            />
                        </Card>
                    </Col>
                
            </Row>
            <Divider/>
            
            <Row justify="center" >
            <Typography.Title level={4}>抓包信息</Typography.Title>
            </Row>
            <Row justify="center">
            <Button enable={linkDumpUrl===""} onClick={()=>{
                StartLinkWebshell(linkID,nodeIndex,(response)=>{
                    const linkWebshellInfo = response.data.data
                    const webShellUrl = `http://${linkWebshellInfo.addr}:${linkWebshellInfo.port}`
                    setLinkDumpUrl(webShellUrl)
                })
            }}>启动抓包</Button>
            </Row>
            <Card style={{width:"90vw",marginLeft:"5vw",marginRight:"5vw"}}>
            <Row justify="center" ><iframe style={{width:"80vw",height:"720px"}} enable={linkDumpUrl!==""} src={linkDumpUrl===""?"":linkDumpUrl}/></Row>
            </Card>
        </div>
    )
}