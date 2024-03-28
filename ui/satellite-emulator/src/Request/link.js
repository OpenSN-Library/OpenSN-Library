import axios from "axios";
import { UrlBase } from "./base";

export const GetLinkList = (callback) => {
    axios.get(UrlBase+"/link/").then((response)=>{
        callback(response)
    })
}

export const GetLinkParameterList = (callback) => {
    axios.get(UrlBase+"/link_parameter/").then((response)=>{
        callback(response)
    })
}