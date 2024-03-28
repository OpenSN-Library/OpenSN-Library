import axios from "axios";
import { UrlBase } from "./base";
/*
{
    "code": 0,
    "msg": "Success",
    "data": {
        "running": false,
        "type_config": {
            "Satellite": {
                "image": "docker.io/realssd/satellite-router:latest",
                "container_envs": null,
                "resource_limit": {
                    "nano_cpu": 40000000,
                    "memory_byte": 67108864
                }
            }
        }
    }
}
*/
export const GetEmulationConfig = (callback) => {
    axios.get(UrlBase+"/emulation/").then((response)=>{
        callback(response)
    })
}

export const StartEmulation = (callback) => {
    axios.post(UrlBase+"/emulation/start").then((response)=>{
        callback(response)
    })
}

export const StopEmulation = (callback) => {
    axios.post(UrlBase+"/emulation/stop").then((response)=>{
        callback(response)
    })
}

export const AddTopology = (topology,callback) => {
    axios.post(UrlBase+"/emulation/topology",topology).then((response)=>{
        callback(response)
    })
}

export const UpdateInstanceType = (instanceTypes,callback) => {
    axios.post(UrlBase+"/emulation/update",instanceTypes).then((response)=>{
        callback(response)
    })
}

export const ResetEmulation = (callback) => {
    axios.post(UrlBase+"/emulation/reset").then((response)=>{
        callback(response)
    })
}