import axios from "axios";
import { UrlBase } from "./base";

export const GetFileList = (relative_path,callback) => {
    axios.get(UrlBase + `/file/list?path=${relative_path}`).then((response) => {
        callback(response);
    });
}

export const UploadFile = (file,relative_path,callback) => {
    const formData = new FormData();
    formData.append("file", file);
    axios.post(UrlBase + `/file/?path=${relative_path}`, formData).then((response) => {
        callback(response);
    });
}

export const DeleteFile = (relative_path,callback) => {
    axios.delete(UrlBase + `/file/?path=${relative_path}`).then((response) => {
        callback(response);
    });
}

export const DownloadFile = (relative_path,callback) => {
    const filename = relative_path.split("/").pop()
    axios.get(UrlBase + `/file/download/${filename}?path=${relative_path}`).then((response) => {
        callback(response);
    });
}

export const PreivewFile = (relative_path,callback) => {
    axios.get(UrlBase + `/file/preview?path=${relative_path}`).then((response) => {
        callback(response);
    });
}