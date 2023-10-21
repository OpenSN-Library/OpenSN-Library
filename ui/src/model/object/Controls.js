import {Camera, Renderer} from "three";
import {OrbitControls} from "three/examples/jsm/controls/OrbitControls";

export const initControls = (camera: Camera, renderer: Renderer, labelRender) => {
  let controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.enableZoom = true;
  controls.autoRotate = false;
  controls.autoRotateSpeed = 2;
  controls.enablePan = true;

  return controls;
}