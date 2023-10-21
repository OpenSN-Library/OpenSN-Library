import {Color, Fog, Scene} from "three";

export const initScene = () => {
  let scene = new Scene();
  scene.background = new Color(0x1d1d1d);
  scene.fog = new Fog(0x020924, 200, 1000);
  window.scene = scene;

  return scene;
}