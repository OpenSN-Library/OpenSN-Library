
import axios from 'axios';
import { UrlBase } from './base';

export function CreateNamespace(nsConfig,callback) {
    console.log(nsConfig)
    axios.post(UrlBase+`/api/namespace/create`, nsConfig).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function GetNamespaceList(callback) {
    axios.get(UrlBase+`/api/namespace/list`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function GetNamespaceDetail(name,callback) {
    axios.get(UrlBase+`/api/namespace/${name}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function UpdateNamespace(name,nsConfig,callback) {
    axios.post(UrlBase+`/api/namespace/${name}/update`, nsConfig).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function DeleteNamespace(name,callback) {
    axios.delete(UrlBase+`/api/namespace/${name}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}


export function StartNamespace(name,callback) {
    axios.post(UrlBase+`/api/namespace/${name}/start`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

export function StopNamespace(name,callback) {
    axios.post(UrlBase+`/api/namespace/${name}/stop`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}