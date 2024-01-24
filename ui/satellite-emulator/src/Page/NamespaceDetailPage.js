import { useEffect,useState} from "react"
import { Namespace } from "../Model/Namespace"
import { GetNamespaceDetail } from "../Request/Namespace"
import { GetNamespacePosition } from "../Request/Position"
import { Typography } from "antd"
import { useParams } from "react-router-dom"
import {Cartesian3} from "cesium"
import { Viewer, Entity } from "resium";


export function NamespaceDetilPage() {
    const [detail, setDetail] = useState(new Namespace())
    const [position,setPositon] = useState({})
    const params = useParams()
    useEffect(() => {
        GetNamespaceDetail(params.name,(response) => {
            setDetail(response.data.data)
            if (detail.running) {
                console.log("start")
                setInterval(() => {
                    GetNamespacePosition(params.name,(response) => {
                        setPositon(response.data.data)
                    })
                },1000)
            }
        })

    }, [params.name])
    return (
        <div>
            <Typography.Title>
                {detail.name}
            </Typography.Title>
        
        {detail.running?<div style={{height:"600px",width:"80%"}}>
            <Viewer>
                {
                    detail.instance_infos.map((item,index) => {

                        return (
                            <Entity
                                key={index}
                                name={item.name}
                                description={JSON.stringify(item.extra, null, '\t')}
                                position={position[item.instance_id]?Cartesian3.fromRadians(
                                    position[item.instance_id].longitude, 
                                    position[item.instance_id].latitude,
                                    position[item.instance_id].height
                                ):Cartesian3.fromRadians(0,0,0)}
                                point={{ pixelSize: 10 }}
                            />
                        )
                    })
                    
                }
            
            </Viewer>
        </div>:<div/>}
            
       </div>
    )
}
