import {CSS3DRenderer} from 'three/examples/jsm/renderers/CSS3DRenderer';

export const initLabelRenderer = (mount) => {
  let labelRenderer = new CSS3DRenderer();
  labelRenderer.setSize(mount.clientWidth, mount.clientHeight);
  labelRenderer.domElement.style.position = "absolute";
  labelRenderer.domElement.style.top = 0;
  labelRenderer.domElement.style.pointerEvents = 'none';

  return labelRenderer;
}