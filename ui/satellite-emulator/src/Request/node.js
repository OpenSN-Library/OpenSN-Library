import axios from 'axios';
import { UrlBase } from './base';

export function GetNodeList(callback) {
    axios.get(UrlBase+`/api/node/list`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function GetNodeDetail(index,callback) {
    axios.get(UrlBase+`/api/node/${index}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}