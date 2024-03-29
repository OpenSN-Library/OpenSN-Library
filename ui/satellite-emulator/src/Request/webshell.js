import axios from "axios";
import { UrlBase } from "./base";

export const StartInstanceWebshell = (nodeIndex,instanceID,callback) => {
    axios.post(UrlBase+"/webshell/instance",{
        node_index:Number(nodeIndex),
        instance_id:instanceID,
        expire_minute:5,
    }).then((response)=>{
        callback(response)
    })
}

export const StartLinkWebshell = (linkID,nodeIndex,callback) => {
    axios.post(UrlBase+"/webshell/link",{
        link_id:linkID,
        node_index:Number(nodeIndex),
        expire_minute:5,
    }).then((response)=>{
        callback(response)
    })
}