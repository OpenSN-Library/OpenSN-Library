import {Box3, BufferGeometry, Group, Line, LineBasicMaterial, LineCurve3, Sphere, Vector3} from "three";
import {sateLocation2xyz} from "./sateLocation2xyz";

export const collideCheck = (sateLocation, group: Group) => {
  /**
   * @description: 碰撞检测, 如果卫星与地面站相交, 则绘制一条红色的连线
   * @param sat_position: {long: number, lat: number, height: number} 卫星的经纬度和高度
   * @param group: Group
   */
  if (!group.getObjectByName("ground_station") || !sateLocation) {
    return;
  }
  group.remove(group.getObjectByName("line_gs"));
  const gs_box = new Box3().setFromObject(group.getObjectByName("ground_station"));

  const xyz = sateLocation2xyz(sateLocation);
  const p2 = new Vector3(xyz.x, xyz.y, xyz.z);
  const cone_sphere = new Sphere(p2, 30);

  if (cone_sphere.intersectsBox(gs_box)) {
    const p1 = gs_box.min.add(gs_box.max).multiplyScalar(0.5);

    let points = new LineCurve3(p1, p2).getPoints(100);
    const geometry = new BufferGeometry();
    geometry.setFromPoints(points);

    const material = new LineBasicMaterial({color: '#ff0000'});
    const line = new Line(geometry, material);
    line.name = "line_gs";
    group.add(line);
  }
}