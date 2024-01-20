import { Button, Drawer, Space} from 'antd';
export const ContainerEnvDrawer = ({dataBuf,setDataBuf,open,setOpen}) => {

  const onClose = () => {
    setOpen(false)
  }

  return (
    <Drawer
        title="Create a new account"
        width={720}
        onClose={onClose}
        open={open}
        styles={{
          body: {
            paddingBottom: 80,
          },
        }}
        extra={
          <Space>
            <Button onClick={onClose}>Cancel</Button>
            <Button onClick={onClose} type="primary">
              Submit
            </Button>
          </Space>
        }
      >

      </Drawer>
  );
};