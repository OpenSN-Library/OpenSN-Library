import axios from "axios";
import { UrlBase } from "./base";

export const GetDatabaseItems = (callback) => {
    axios.get(UrlBase+"/database/items").then((response)=>{
        callback(response)
    })
}

export const UpdateDatabaseItem = (data,callback) => {
    axios.post(UrlBase+"/database/update",data).then((response)=>{
        callback(response)
    })
}

export const DeleteDatabaseItem = (data,callback) => {
    axios.post(UrlBase+"/database/delete",data).then((response)=>{
        callback(response)
    })
}