import {AmbientLight, DirectionalLight, HemisphereLight, Scene} from 'three'

export const initLight = (scene: Scene) => {
  let ambientLight = new AmbientLight(0xcccccc, 1.1);
  scene.add(ambientLight);

  let directionalLight = new DirectionalLight(0xffffff, 0.2);
  directionalLight.position.set(1, 0.1, 0).normalize();
  scene.add(directionalLight);

  directionalLight = new DirectionalLight(0xff2ffff, 0.2);
  directionalLight.position.set(1, 0.1, 0.1).normalize();
  scene.add(directionalLight);

  let hemiLight = new HemisphereLight(0xffffff, 0x444444, 0.2);
  hemiLight.position.set(0, 1, 0);
  scene.add(hemiLight);

  directionalLight = new DirectionalLight(0xffffff);
  directionalLight.position.set(1, 500, -20);
  directionalLight.castShadow = true;
  directionalLight.shadow.camera.top = 18;
  directionalLight.shadow.camera.bottom = -10;
  directionalLight.shadow.camera.left = -52;
  directionalLight.shadow.camera.right = 12;
  scene.add(directionalLight);
}