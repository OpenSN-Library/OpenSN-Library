import {useEffect} from "react";
import {socket} from '../utils/WebSocket';
import './VideoTransmitPage.css';

const VideoTransmitPage = (props) => {


  useEffect(() => {
    const srcWebSocket = socket.init(props.ports['src_port']);
    const tcpWebSocket = socket.init(props.ports['tcp_port']);
    const myWebSocket = socket.init(props.ports['my_port']);

    return () => {
      if (srcWebSocket) {
        srcWebSocket.close();
      }
      if (tcpWebSocket) {
        tcpWebSocket.close();
      }
      if (myWebSocket) {
        myWebSocket.close();
      }
    }
  }, [props.ports]);


  return (
      <div>
        <div className={'videoPlayerSrc'}>
          <h2 className={'videoTitle'}>源视频文件</h2>
          <img id={'videoPlayer' + props.ports['src_port']} alt={'源视频文件'}/>
        </div>
        <div className={'videoPlayerTcpDst'}>
          <h2 className={'videoTitle'}>TCP协议传输后视频文件</h2>
          <img id={'videoPlayer' + props.ports['tcp_port']} alt={'TCP协议传输后视频文件'}/>
        </div>
        <div className={'videoPlayerMyDst'}>
          <h2 className={'videoTitle'}>新协议传输后视频文件</h2>
          <img id={'videoPlayer' + props.ports['my_port']} alt={'新协议传输后视频文件'}/>
        </div>
      </div>
  )
};

export default VideoTransmitPage;