import axios from "axios";
import { UrlBase } from "./base";

export const GetAllNodeResource = (callback) => {
    axios.get(UrlBase+`/resource/last/node/all`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export const GetAllLinkLastResource = (callback) => {
    axios.get(UrlBase+`/resource/last/link/all`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export const GetAllInstanceLasteResource = (callback) => {
    axios.get(UrlBase+`/resource/last/instance/all`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export const GetPeriodNodeResource = (node_id,period,callback) => {
    axios.get(UrlBase + `/resource/period/node/${node_id}?period=${period}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export const GetPeriodInstanceResource = (instance_id,period,callback) => {
    axios.get(UrlBase + `/resource/period/instance/${instance_id}?period=${period}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export const GetPeriodLinkResource = (link_id,period,callback) => {
    axios.get(UrlBase + `/resource/period/link/${link_id}?period=${period}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}