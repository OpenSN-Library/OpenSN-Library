import {Group} from "three";
import {GLTFLoader} from "three/examples/jsm/loaders/GLTFLoader";
import {DRACOLoader} from "three/examples/jsm/loaders/DRACOLoader";
import gs_3d from "../../3d_object/satellite_ground_station.glb";
import para from "../../parameters.json"
import {DoubleSide, Mesh, MeshBasicMaterial, PlaneGeometry, TextureLoader, Vector3} from "three";
import wave_img from "../../img/wave.png";

export const GroundStation = (position) => {
  /**
   * @description 根据地面位置初始化地面站
   * @param position {long: float, lat: float} 地面站位置
   * @return Group 地面站模型组
   */
  const group = new Group();
  const loader = new GLTFLoader();
  const dracoLoader = new DRACOLoader();
  dracoLoader.setDecoderPath('/draco/');
  dracoLoader.preload(gs_3d);
  loader.setDRACOLoader(dracoLoader);

  const rad = para.earth_radius;
  const theta = Math.PI / 2 - position.lat;
  const phi = - position.long;
  const x = rad * Math.sin(theta) * Math.cos(phi);
  const y = rad * Math.cos(theta);
  const z = rad * Math.sin(theta) * Math.sin(phi);
  loader.load(gs_3d, (gltf) => {
    const scale = 0.2;
    gltf.scene.scale.set(scale, scale, scale);
    gltf.scene.position.set(x, y, z);
    gltf.scene.rotateY(-phi);
    gltf.scene.rotateZ(-theta - Math.PI / 4);
    gltf.scene.name = "ground_station";
    group.add(gltf.scene);
  }, undefined, (err) => {
    console.log(err);
  });

  group.add(wave(x * 1.1, y * 1.1 - 2, z * 1.1));
  return group;
}

const wave = (x, y, z) => {
  const plane = new PlaneGeometry(1, 1);
  const texture = new TextureLoader().load(wave_img);
  const material = new MeshBasicMaterial({
    color: 0x22ffcc,
    map: texture,
    transparent: true,
    opacity: 1,
    side: DoubleSide,
    depthWrite: false
  });

  const wave_mesh = new Mesh(plane, material);
  wave_mesh.position.set(x, y, z);
  const gs_normal_vec = new Vector3(x, y, z).normalize();
  const mesh_normal_vec = new Vector3(0, 0, 1);
  wave_mesh.quaternion.setFromUnitVectors(mesh_normal_vec, gs_normal_vec);
  wave_mesh.size = 10;
  wave_mesh._s = Math.random() + 1.0;

  wave_mesh.name = "wave";
  return wave_mesh;
}

