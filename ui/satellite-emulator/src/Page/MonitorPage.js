import { Viewer,Entity,PointGraphics } from "resium"
import { Cartesian3 } from "cesium"

export const MonitorPage = () => {

    return (

        <Viewer infoBox={false}>
        <Entity position={Cartesian3.fromDegrees(139.767052, 35.681167, 100)} name="Tokyo" description="Hello, world.">
        <PointGraphics pixelSize={10} />
        </Entity>
        </Viewer>
    
        )
    }