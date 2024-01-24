import { Button, Divider, Drawer, Input, Space, Typography} from 'antd';
import { useState } from 'react';
import { ResourceLimit } from '../Model/Namespace';
export const InstanceTypeDrawer = ({dataBuf,setDataBuf,open,setOpen}) => {

  const onClose = () => {
    setOpen(false)
  }

  const onConfirm = () => {
    dataBuf.ns_config.image_map[typeName] = imageName
    dataBuf.ns_config.resource_map[typeName] = resourceLimit
    setTypeName("")
    setImageName("")
    setResourceLimit(new ResourceLimit())
    setDataBuf(dataBuf)
    setOpen(false)
    console.log(dataBuf)
  }

  const [typeName,setTypeName] = useState("")
  const [imageName,setImageName] = useState("")
  const [resourceLimit,setResourceLimit] = useState(new ResourceLimit())

  return (
    <Drawer
        title="添加实例类型"
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
          <Typography.Text strong>实例类型: </Typography.Text>
          <Input type="text" style={{width:"200px"}} onChange={(e) => {
              setTypeName(e.target.value);
          }} />
        </Typography.Paragraph>
        <Divider/>
        <Typography.Paragraph>
          <Typography.Text strong>镜像名称: </Typography.Text>
          <Input type="text" style={{width:"200px"}} onChange={(e) => {
              setImageName(e.target.value);
          }} />
        </Typography.Paragraph>
        <Divider/>
        <Typography.Paragraph>
          <div>
            <Typography.Text strong>CPU限额(1e-9 Core): </Typography.Text>
            <Input type="text" style={{width:"200px"}} onChange={(e) => {
                resourceLimit.nano_cpu = e.target.value;
                setResourceLimit(resourceLimit);
            }} />
          </div>
          <div>
            <Typography.Text strong>内存限额(Byte): </Typography.Text>
            <Input type="text" style={{width:"200px"}} onChange={(e) => {
                resourceLimit.memory_byte = e.target.value;
                setResourceLimit(resourceLimit);
            }} />
          </div>
        </Typography.Paragraph>
        <Typography.Paragraph>
          "限额必须为整数, 可以以[K|M|G|T]为结尾单位, 不区分大小写"
        </Typography.Paragraph>
      </Drawer>
  );
};