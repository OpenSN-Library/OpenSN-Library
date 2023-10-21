import {Group, PerspectiveCamera, Raycaster, Renderer, Scene, Vector2,} from "three";
import {OrbitControls} from "three/examples/jsm/controls/OrbitControls";
import {Component} from "react";

import {initRenderer} from "./object/Renderer";
import {initLabelRenderer} from "./object/LabelRender";
import {initScene} from "./object/Scene";
import {initCamera} from "./object/Camera";
import {initControls} from "./object/Controls";
import {initEarth} from "./object/Earth";
import {initBackground} from "./object/Background";
import {initLight} from "./object/Light";
import {addLinesBetweenSatellites, addLinesBetweenSatelliteAndGroundStation} from "./util/addLines";
import {sateLocation2xyz} from "./util/sateLocation2xyz";
import {Cone} from "./object/Cone";
// import {collideCheck} from "./util/collideCheck";
import {waveAnimate} from "./util/waveAnimate";
import {addRoutePath} from "./util/addRoutePath";
import {addSatellites} from "./util/addSatelites";
import {addGroundStations} from "./util/addGroundStations";
import para from "../parameters.json";

class Model extends Component {
  camera: PerspectiveCamera;
  renderer: Renderer;
  scene: Scene;
  controls: OrbitControls;

  constructor(props) {
    super(props);
    this.state = {
      group: new Group(),
    };
  }

  init() {
    // 初始化场景、相加、渲染器、控制器
    this.scene = initScene();
    this.camera = initCamera(this.mount);
    this.renderer = initRenderer(this.mount);
    this.labelRender = initLabelRenderer(this.mount);
    this.controls = initControls(this.camera, this.renderer);

    this.scene.add(this.state.group);
    this.mount.appendChild(this.renderer.domElement);
    this.mount.appendChild(this.labelRender.domElement);

    // 初始化灯光、地球、地面站
    initBackground(this.scene);
    initLight(this.scene);
    initEarth(this.state.group);
    this.animate();

    // 添加鼠标点击事件，获取鼠标点击的卫星
    this.mount.addEventListener("click", this.onMouseOver, false)
  }

  onMouseOver = (event) => {
    let intersects = this.getIntersects(event);

    // 筛选点击的物体，如果点击的物体是卫星，则将选择的卫星设置为该卫星（定义在父组件）
    if (intersects.length > 0) {
      let object = null;
      intersects.forEach((element) => {
        if (element.object.name !== 'longLine') {
          object = element.object;
        }
      })
      while (object !== null && !object.name.startsWith(para.satellite_name)) {
        object = object.parent;
        if (object === null) break;
      }
      if (object) {
        console.log(object.name)
        this.props.displayPara(object.name);
      }
    }
  }

  getIntersects = (event) => {
    /**
     * @description: 使用Raycaster获取鼠标点击直线上的所有物体
     * return：返回一个数组，数组中包含所有与直线相交的物体
     */
    event.preventDefault();

    let rayCaster = new Raycaster();
    let mouse = new Vector2();
    mouse.x = ((event.clientX - 260) / (this.mount.clientWidth)) * 2 - 1;
    mouse.y = -(event.clientY / this.mount.clientHeight) * 2 + 1;
    console.log(mouse)

    rayCaster.setFromCamera(mouse, this.camera);
    return rayCaster.intersectObjects(this.state.group.getObjectByName("sateGroup").children, true);
  }

  animate = () => {
    requestAnimationFrame(this.animate);
    if (this.controls) {
      this.controls.update();
    }
    const groundGroup = this.state.group.getObjectByName('groundGroup');
    if (groundGroup) {
      waveAnimate(groundGroup);
    }
    this.renderer.render(this.scene, this.camera);
    this.labelRender.render(this.scene, this.camera);
  }

  update = () => {
    /**
     * @description: 更新位置
     * @description: 卫星、标签 直接更新位置
     * @description: 卫星连线、路由路径 删除原有的方式，重新添加新的轨迹和路径
     */
    let sateGroup = this.state.group.getObjectByName("sateGroup");
    if (sateGroup && sateGroup.children.length > 0) { // 如果卫星组存在,直接更新卫星位置
      // 更新卫星位置
      sateGroup.children.forEach((group) => {
        const name = group.name;
        const sate = group.getObjectByName('sate');
        const longLine = group.getObjectByName('longLine');
        const sateLocation = this.props.sateLocations[name];
        if (!(sateLocation["long"] === 0 && sateLocation["lat"] === 0 && sateLocation["height"] === 0) && sate && longLine) {
          const xyz = sateLocation2xyz(sateLocation);
          sate.position.set(xyz.x, xyz.y, xyz.z);
          longLine.visible = this.props.flags["longLine"];
        }
      })
      // 更新卫星标签位置
      this.state.group.getObjectByName("labelGroup").children.forEach((element) => {
        const sateLocation = this.props.sateLocations[element.name];
        if (!(sateLocation["long"] === 0 && sateLocation["lat"] === 0 && sateLocation["height"] === 0)) {
          let xyz = sateLocation2xyz(sateLocation)
          element.position.set(xyz.x, xyz.y, xyz.z);
          element.visible = this.props.flags["label"];
        }
      })
      // 更新卫星辐射范围圆锥，删除原有的圆锥，重新添加新的圆锥
      this.state.group.remove(this.state.group.getObjectByName("cone"));
      this.state.group.add(Cone(this.props.sateLocations[this.props.selected_sate]));
      // 碰撞检测，卫星与地面站碰撞时添加连线
      // collideCheck(this.props.sateLocations[this.props.selected_sate], this.state.group);
    } else { // 如果卫星组不存在，添加卫星
      // 删除原有的卫星组（此时卫星组内没有子元素）
      this.state.group.remove(this.state.group.getObjectByName("sateGroup"));
      this.state.group.remove(this.state.group.getObjectByName("groundGroup"));
      addSatellites(this.props.sateLocations, this.state.group);
      addGroundStations(this.props.groundPosition, this.state.group);
    }

    // 删除卫星连线、路由路径
    this.state.group.remove(this.state.group.getObjectByName("lineGroup"));
    this.state.group.remove(this.state.group.getObjectByName("pathGroup"));

    // 重新添加卫星连线、路由路径
    if (this.props.flags["connect"]) {
      addLinesBetweenSatellites(this.state.group, this.props.sateLocations, this.props.sateConnections);
      addLinesBetweenSatelliteAndGroundStation(
          this.state.group, this.props.sateLocations, this.props.groundPosition, this.props.groundConnections);
    }
    addRoutePath(this.state.group, this.props.routePath, this.props.sateLocations);
  }

  componentDidMount() {
    this.init();
    window.addEventListener('resize', this.onWindowResize);
    setTimeout(() => this.onWindowResize(), 10);
  }

  componentWillUnmount() {
    this.mount.removeChild(this.renderer.domElement);
    window.removeEventListener('resize', this.onWindowResize);
  }

  onWindowResize = () => {
    this.camera.aspect = this.mount.clientWidth / this.mount.clientHeight;
    this.camera.updateProjectionMatrix();

    this.renderer.setSize(this.mount.clientWidth, this.mount.clientHeight);
    this.renderer.render(this.scene, this.camera);

    this.labelRender.setSize(this.mount.clientWidth, this.mount.clientHeight);
    this.labelRender.render(this.scene, this.camera);
  }

  render() {
    this.update();
    return (<div
        id="canvas"
        style={{width: '100%', height: '100%'}}
        ref={(mount) => {
          this.mount = mount
        }}
    />);
  }
}

export default Model;