import { useEffect,useState} from "react"
import { Namespace } from "../Model/Namespace"
import { GetNamespaceDetail } from "../Request/Namespace"
import { GetNamespacePosition } from "../Request/Position"
import { Card, Typography } from "antd"
import { useParams } from "react-router-dom"
import {Cartesian3,Color} from "cesium"
import { CesiumWidget, Entity, PolylineGraphics, Viewer } from "resium";


export function NamespaceDetilPage() {
    const [detail, setDetail] = useState(new Namespace())
    const [position,setPositon] = useState({})
    const [updateValid,setUpdateValid] = useState(false)
    const params = useParams()
    useEffect(() => {
        GetNamespaceDetail(params.name,(response) => {
            setDetail(response.data.data)
            console.log("start2")
            if (response.data.data.running) {
                console.log("start3")
                setInterval(() => {
                    GetNamespacePosition(params.name,(response) => {
                        setPositon(response.data.data)
                    })
                },5000)
            }
        })
        console.log("start1")
    }, [])

    return (
        <div>
            <Typography.Title>
                {detail.name}
            </Typography.Title>
        
        {detail.running ?<Card style={{width:"100%"}}>
            <Viewer timeline={false} homeButton={false} geocoder={false} animation={false} navigationHelpButton={false} fullscreenButton={false}>
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
                {
                    detail.link_infos.map((item,index) => {
                        console.log(Object.keys(position).length)
                        if (item.parameter["connect"]===1 && position[item.connect_instance[0]]!==undefined && position[item.connect_instance[1]]!=undefined) {
                            
                            const array = [
                                position[item.connect_instance[0]].longitude, 
                                position[item.connect_instance[0]].latitude,
                                position[item.connect_instance[0]].height,
                                position[item.connect_instance[1]].longitude, 
                                position[item.connect_instance[1]].latitude,
                                position[item.connect_instance[1]].height,
                            ]
                            return (
                                <Entity key={item.link_id} name="PolygonGraphics" description="PolygonGraphics!!" >
                                    <PolylineGraphics
                                        positions={Cartesian3.fromRadiansArrayHeights(array)}
                                        width={1}
                                        material={Color.CYAN}
                                    />
                                </Entity>
                            )
                        } else {
                            return null
                        }
                    })

                }
                
            </Viewer>
        </Card>:<div/>}

       </div>
    )
}
