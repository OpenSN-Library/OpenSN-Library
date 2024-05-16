import { useEffect,useState } from "react";
import { GetDatabaseItems } from "../Request/databse";
import { Button, Card, Col, Divider, Row, Typography,Tree } from "antd";
import {CarryOutOutlined} from '@ant-design/icons/CarryOutOutlined';
import { VanillaJSONEditor } from "../Component/VanillaJSONEditor";
import { DeleteDatabaseItem,UpdateDatabaseItem } from "../Request/databse";

export const DatabasePage = () => {

    const [showLine, setShowLine] = useState(true);
    const [showIcon, setShowIcon] = useState(false);
    const [showLeafIcon, setShowLeafIcon] = useState(true);
    const [databaseItems,setDatabaseItems] = useState({})
    const [treeKeys,setTreeKeys] = useState([])
    const [selectKey,setSelectKey] = useState("/")
    const [editorContent,setEditorContent] = useState({
        json:{},
        text:""
    })
    useEffect(()=>{
        GetDatabaseItems((response)=>{
            const data = response.data.data
            const tmpTreeKeys = {}
            const items = {}
            Object.keys(data).forEach((item)=>{
                const keys = item.split("/")
                let tmp = tmpTreeKeys
                keys.forEach((key)=>{
                    if(key===""){
                        return
                    }
                    if(tmp[key]===undefined){
                        tmp[key] = {
                            realKey : keys.slice(0,keys.indexOf(key)+1).join("/"),
                            children:{}
                        }
                    }
                    tmp = tmp[key].children
                })
                items[item] = data[item]
            })
            console.log(tmpTreeKeys)
            setTreeKeys(tmpTreeKeys)
            setDatabaseItems(items)
        })
    },[])

    const transTreeData = (data) => {
        const res = []
        Object.keys(data).forEach((key)=>{
            const item = data[key]
            res.push({
                key:data[key].realKey,
                title:key,
                children:transTreeData(item.children)
            })
        })
        return res
    }

    const onSelect = (selectedKeys, info) => {
        console.log('selected', selectedKeys, info);
        setSelectKey(selectedKeys[0])
        setEditorContent({
            json:databaseItems[selectedKeys[0]]?JSON.parse(databaseItems[selectedKeys[0]]):{},
            text:undefined
        })
    }


    return (
        <Card style={{height:"85vh"}}>
            <Row justify="space-between" style={{height:"100%"}} >
                <Col style={{height:"100%"}}>
                    <Card style={{width:"20vw",height:"80vh"}}>
                        <Typography.Title level={4}>数据库键列表</Typography.Title>
                        <Tree
                            style={{height:"65vh",overflowY:"auto"}}
                            showLine={showLine ? { showLeafIcon } : false}
                            showIcon={showIcon}
                            defaultExpandedKeys={['/']}
                            onSelect={onSelect}
                            treeData={transTreeData(treeKeys)}
                        />
                    </Card>
                </Col>
                <Col style={{height:"100%"}}>
                
                <Card style={{width:"70vw",height:"80vh"}}>
                <Typography.Title level={4}>数据库条目</Typography.Title>
                    <VanillaJSONEditor
                        style={{height:"70vh"}}
                        content={editorContent}
                        readOnly={false}
                        onChange={(content) => {}}
                    />
                    <Divider/>
                    <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                        onClick={()=>{
                            UpdateDatabaseItem({
                                key:selectKey,
                                value:editorContent.text
                            },(response)=>{
                                console.log(response)
                            })
                        }}
                    >提交</Button>
                    <Button type="primary" style={{marginRight:"10px",marginLeft:"10px"}}
                        onClick={()=>{
                            DeleteDatabaseItem({
                                key:selectKey
                            },(response)=>{
                                console.log(response)
                            })
                        }}
                    >删除</Button>
                </Card>
                </Col>
            </Row>
        </Card>
    );
    }