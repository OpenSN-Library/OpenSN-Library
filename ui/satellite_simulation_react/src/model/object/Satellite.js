import {Group} from "three";
import {GLTFLoader} from "three/examples/jsm/loaders/GLTFLoader";
import {DRACOLoader} from "three/examples/jsm/loaders/DRACOLoader";
import sate_3d from "../../3d_object/low_poly_satellite.glb";
import {addLongLine} from "../util/addLongLine";
import {sateLocation2xyz} from "../util/sateLocation2xyz";

export const Satellite = (sateLocation, name) => {
  /**
   * @description 根据卫星位置信息，创建卫星模型、标签、卫星经线
   * @param sateLocation 卫星位置信息
   * @param name 卫星名称
   * @returns {Group} 卫星模型、标签、卫星经线组
   */
  let sate = new Group();
  let xyz = sateLocation2xyz(sateLocation);

  const loader = new GLTFLoader();
  const dracoLoader = new DRACOLoader();
  dracoLoader.setDecoderPath('/draco/');
  dracoLoader.preload(sate_3d);
  loader.setDRACOLoader(dracoLoader);

  loader.load(sate_3d, (gltf) => {
    gltf.scene.scale.set(0.5, 0.5, 0.5);
    gltf.scene.position.set(xyz.x, xyz.y, xyz.z);
    gltf.scene.name = "sate";
    sate.add(gltf.scene);
  }, undefined, (err) => {
    console.log(err);
  });

  let line = addLongLine(sateLocation);
  line.name = 'longLine';
  sate.add(line);
  return sate;
}
