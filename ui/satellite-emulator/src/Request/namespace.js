
import axios from 'axios';

function CreateNamespace(nsConfig,callback) {
    axios.post('/api/namespace/create',json = nsConfig).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function GetNamespaceList(callback) {
    axios.get(`/api/namespace/list`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function GetNamespaceDetail(name,callback) {
    axios.get(`/api/namespace/${name}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function UpdateNamespace(name,nsConfig,callback) {
    axios.post(`/api/namespace/${name}/update`,json = nsConfig).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function DeleteNamespace(name,callback) {
    axios.delete(`/api/namespace/${name}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}


function StartNamespace(name,callback) {
    axios.post(`/api/namespace/${name}/start`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function StopNamespace(name,callback) {
    axios.post(`/api/namespace/${name}/stop`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

exports.CreateNamespace = CreateNamespace;
exports.GetNamespaceList = GetNamespaceList;
exports.GetNamespaceDetail = GetNamespaceDetail;
exports.UpdateNamespace = UpdateNamespace;
exports.DeleteNamespace = DeleteNamespace;
exports.StartNamespace = StartNamespace;
exports.StopNamespace = StopNamespace;