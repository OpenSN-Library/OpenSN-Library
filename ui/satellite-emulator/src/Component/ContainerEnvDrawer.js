import { Button, Drawer, Input, Space, Typography} from 'antd';
import { useState } from 'react';
export const ContainerEnvDrawer = ({dataBuf,setDataBuf,open,setOpen}) => {

  const onClose = () => {
    setOpen(false)
  }

  

  const [envKey,setEnvKey] = useState("")
  const [envValue,setEnvValue] = useState("")

  const onConfirm = () => {
    setEnvKey("")
    setEnvValue("")
    dataBuf.ns_config.container_envs[envKey] = envValue
    setDataBuf(dataBuf)
    setOpen(false)
    console.log(dataBuf)
  }

  return (
    <Drawer
        title="添加初始化环境变量"
        width={400}
        onClose={onClose}
        open={open}
        styles={{
          body: {
            paddingBottom: 80,
          },
        }}
        extra={
          <Space>
            <Button onClick={onClose}>取消</Button>
            <Button onClick={onConfirm} type="primary">
              确认
            </Button>
          </Space>
        }
      >
        <Typography.Paragraph>
          <Typography.Text strong>环境变量Key: </Typography.Text>
          <Input type="text" style={{width:"200px"}} onChange={(e) => {
              setEnvKey(e.target.value)
          }} />
        </Typography.Paragraph>
        <Typography.Paragraph>
          <Typography.Text strong>环境变量Value: </Typography.Text>
          <Input type="text" style={{width:"200px"}} onChange={(e) => {
              setEnvValue(e.target.value)
          }} />
        </Typography.Paragraph>
      </Drawer>
  );
}