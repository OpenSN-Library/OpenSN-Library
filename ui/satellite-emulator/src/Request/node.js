import axios from 'axios';

function GetNodeList(callback) {
    axios.get(`/api/node/list`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

function GetNodeDetail(index,callback) {
    axios.get(`/api/node/${index}`).then(function (response) {
        callback(response);
    }).catch(function (error) {
        console.error(error);
    });
}

exports.GetNodeList = GetNodeList;
exports.GetNodeDetail = GetNodeDetail;