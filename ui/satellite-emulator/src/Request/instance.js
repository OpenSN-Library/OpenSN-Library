import axios from "axios"
import { UrlBase } from "./base"

export const GetInstanceList = (callback) => {
    axios.get(UrlBase+"/instance/").then((response)=>{
        callback(response)
    })
}

export const GetInstanceDetail = (node_index, instance_id,callback) => {
    axios.get(UrlBase+`/instance/${node_index}/${instance_id}`).then((response)=>{
        callback(response)
    })
}