import axios from "axios";
import { UrlBase } from "./base";

export function GetAllPosition(callback) {
    axios.get(UrlBase+`/position/all`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}