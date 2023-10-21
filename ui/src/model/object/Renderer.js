import {WebGLRenderer} from "three";

export const initRenderer = (mount) => {
  let renderer = new WebGLRenderer({
    antialias: true,
    alpha: true
  });
  renderer.setSize(mount.clientWidth, mount.clientHeight);

  return renderer;
}