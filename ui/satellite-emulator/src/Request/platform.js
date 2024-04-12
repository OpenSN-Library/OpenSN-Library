import axios from "axios";
import { UrlBase } from "./base";

export const GetCodeServerConfiguration = (callback) => {
    axios.get(UrlBase+"/platform/address/code_server").then((response)=>{
        callback(response)
    })
}