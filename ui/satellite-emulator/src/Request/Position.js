import axios from "axios";
import { UrlBase } from "./base";

export function GetNamespacePosition(name,callback) {
    axios.get(UrlBase+`/api/position/${name}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}