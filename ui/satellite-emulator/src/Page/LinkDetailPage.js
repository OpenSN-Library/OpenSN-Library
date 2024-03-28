
import { Card, Col, Descriptions, Divider, List, Row, Typography, Badge, Button,Tabs} from "antd"
import { useEffect, useState } from "react"
import { useParams } from "react-router-dom"
import ReactECharts from 'echarts-for-react'
import { GetPeriodInstanceResource } from "../Request/metrics"
import { GetInstanceDetail } from "../Request/instance"

export const LinkDetailPage = () => {
    // const linkID = useParams().link_id
    // const nodeIndex = useParams().node_index
    // const [linkResource,setLinkResource] = useState({})
    // const [resourcePeriod,setResourcePeriod] = useState("10m")
    // const [linkInfo,setLinkInfo] = useState({})

    // useEffect(()=>{
        
    // },[])
    
    // const descriptionDetail = [
    //     {
    //         key:"link_id",
    //         label:"链路ID",
    //         children:linkID,
    //     },
    //     {
    //         key:"link_enable",
    //         label:"链路启用状态",
    //         children:linkInfo.enable?
    //             <Badge status="success" text="已启动" />:
    //             <Badge status="default" text="未启动" />    
    //         ,
    //     },
    //     {
    //         key:"link_type",
    //         label:"链路类型",
    //         children:linkInfo.type,
    //     },
    //     {
    //         key:"node_index",
    //         label:"链路部署机器编号",
    //         children:nodeIndex,
    //     },
    //     {
    //         key:"link_parameter",
    //         label:"链路参数",
    //         children:<div>
    //             <p>
    //                 CPU限额(Nano): {linkInfo?.resource_limit?.nano_cpu}
    //             </p>
    //             <p>
    //                 内存(Byte): {linkInfo?.resource_limit?.memory_byte}
    //             </p>
    //         </div>,
    //     },
    //     {
    //         key:"link_end_infos",
    //         label:"链路两端信息",
    //         children:<List>
    //             {linkInfo.end_infos?Object.keys(linkInfo.end_infos).map((item,index)=>{
    //                 return (
    //                     null
    //                 )
    //             }):[]}
    //         </List>,
    //     },
    // ]

    // const cpuChartOption = {
    //     title: {
    //         text: '链路收发吞吐量(Bps)'
    //     },
    //     tooltip: {},
    //     legend: {
    //         data:['接收数据','发送数据']
    //     },
    //     xAxis: {
    //         data: linkResource[linkID]?linkResource[linkID].map((item)=>item.Time):[]
    //     },
    //     yAxis: {},
    //     series: [{
    //         name: 'CPU使用率',
    //         type:'line',
    //         data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.cpu_usage):[]
    //     }]
    // }

    // const memChartOption = {
    //     title: {
    //         text: '链路数据包统计'
    //     },
    //     tooltip: {},
    //     legend: {
    //         data:['内存使用量','交换内存使用量']
    //     },
    //     xAxis: {
    //         data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.Time):[]
    //     },
    //     yAxis: {},
    //     series: [
    //         {
    //             name: '内存使用量',
    //             type:'line',
    //             data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.mem_byte):[]
    //         },
    //         {
    //             name: '交换内存使用量',
    //             type:'line',
    //             data: instanceResource[instanceID]?instanceResource[instanceID].map((item)=>item.swap_mem_byte):[]
    //         }
    //     ]
        
    // }

    // return (
    //     <div>
    //         <Row justify="center">
    //             <Col>
    //                 <Typography.Title level={2} >实例详情:{instanceID}</Typography.Title>
    //             </Col>
    //         </Row>
    //         <Row justify="center">
    //             <Card>
    //                 <Col>
    //                 <Descriptions bordered items={descriptionDetail} />
    //                 </Col>
    //             </Card>
                
    //         </Row>
    //         <Divider/>
    //         <Row justify="space-between">
                
    //                 <Col style={{marginLeft:"5vw"}}>
    //                     <Card>
    //                         <ReactECharts
    //                             option={cpuChartOption}
    //                             style={{ height: "30vh",width:"40vw" }}
    //                         />
    //                     </Card>
    //                 </Col>
    //                 <Col style={{marginRight:"5vw"}}>
    //                     <Card>
    //                         <ReactECharts
    //                             option={memChartOption}
    //                             style={{ height: "30vh",width:"40vw" }}
    //                         />
    //                     </Card>
    //                 </Col>
                
    //         </Row>
    //         <Divider/>
            
    //         <Row justify="center" >
    //         <Typography.Title level={4}>WebShell</Typography.Title>
    //         <Card style={{width:"90vw",marginLeft:"5vw",marginRight:"5vw"}}>
    //         <Row justify="center" ><iframe style={{width:"80vw",height:"720px"}} src="http://10.134.148.56:8079"/></Row>
    //         </Card>
    //         </Row>
    //     </div>
    // )
}