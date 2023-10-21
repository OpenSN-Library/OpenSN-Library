import {CylinderGeometry, DoubleSide, Mesh, MeshBasicMaterial} from "three";

export const Cone = (position) => {
  /**
   * @description: 为卫星添加辐射范围圆锥
   * @param position: 卫星位置
   * @param group: Group
   */
  if (!position) {
    return;
  }
  const rad = position.height / 63782 + 100;
  const theta = Math.PI / 2 - position.lat;
  const phi = - position.long;
  const x = rad * Math.sin(theta) * Math.cos(phi);
  const y = rad * Math.cos(theta);
  const z = rad * Math.sin(theta) * Math.sin(phi);

  const geometry = new CylinderGeometry(0, 40, rad * 2 / 5, 32, 6, true);
  const material = new MeshBasicMaterial({
    color: 0xffffff,
    side: DoubleSide,
    transparent: true,
    opacity: 0.3,
    depthWrite: false,
  });

  const cylinder = new Mesh(geometry, material);
  cylinder.position.set(x * 4 / 5, y * 4 / 5, z * 4 / 5);
  cylinder.rotateY(-phi);
  cylinder.rotateZ(-theta);
  cylinder.name = "cone";

  return cylinder;
}