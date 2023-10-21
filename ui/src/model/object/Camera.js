import {PerspectiveCamera} from "three";
import para from '../../parameters.json';

export const initCamera = (mount) => {
  let camera = new PerspectiveCamera(60, mount.clientWidth / mount.clientHeight, 1, 2000)
  camera.position.set(0, 0, para.camera_distance);
  camera.lookAt(0, 0, 0);

  return camera;
}