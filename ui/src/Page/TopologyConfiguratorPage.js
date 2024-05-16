import { Row, Typography } from "antd"
import { useState, useEffect } from "react"
import { GetCodeServerConfiguration } from "../Request/platform"
export const TopologyConfiguratorPage = () => {

    const [codeServerUrl,setCodeServerUrl] = useState("")
    useEffect(()=>{
        GetCodeServerConfiguration((response)=>{
            setCodeServerUrl(response.data.data.address!==""?`http://${response.data.data.address}:${response.data.data.port}`:"")
        })
    },[])
    return (
       <Row justify="center">
        {codeServerUrl===""?<Typography.Title>未启动Web配置编辑器</Typography.Title>: <iframe  style={{width:"96vw",height:"84vh"}} src={codeServerUrl}/>}
       
       </Row>
    )
}