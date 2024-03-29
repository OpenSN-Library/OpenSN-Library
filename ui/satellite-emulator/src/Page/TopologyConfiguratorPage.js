import { Row, Typography } from "antd"
import { useState, useEffect } from "react"
export const TopologyConfiguratorPage = () => {

    const [codeServerUrl,setCodeServerUrl] = useState("")
    useEffect(()=>{
        // GetCodeServerUrl((response)=>{
        //     setCodeServerUrl(response.data.data)
        // })
    },[])
    return (
       <Row justify="center">
        {codeServerUrl===""?<Typography.Title>未启动Web配置编辑器</Typography.Title>: <iframe  style={{width:"96vw",height:"84vh"}} src={codeServerUrl}/>}
       
       </Row>
    )
}