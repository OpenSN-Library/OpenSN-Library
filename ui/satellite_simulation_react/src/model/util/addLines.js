import {
  BufferGeometry,
  // CatmullRomCurve3,
  Group,
  Line,
  LineBasicMaterial,
  LineCurve3,
  // Mesh,
  // MeshPhongMaterial,
  // SphereGeometry,
  Vector3
} from "three";
import {sateLocation2xyz} from "./sateLocation2xyz";

export const addLinesBetweenSatellites = (group, sateLocations, sateConnections) => {
  /**
   * 卫星之间的连接状态连线
   * @param group: Group 连线所在组，用于添加到场景中
   * @param sateLocations: {sate_id: {long: float, lat: float, height: float}}
   * @param sateConnections: {sate_id: [sate_id, sate_id, ...]}
   */
  if (!sateLocations || !sateConnections) {
    return;
  }
  const lineGroup = new Group();
  for (let key in sateConnections) {
    let start_position = sateLocations[key];
    let values = sateConnections[key];
    if (start_position.open) {
      for (let i = 0; i < values.length; i++) {
        addLine(lineGroup, start_position, sateLocations[values[i]], '#ffd700');
      }
    } else {
      addLine(lineGroup, start_position, sateLocations[values[0]], '#ffd700');
    }
  }
  lineGroup.name = "lineGroup";
  group.add(lineGroup);
}

export const addLinesBetweenSatelliteAndGroundStation = (group, sateLocations, groundPositions, groundConnections) => {
  /**
   * 卫星与地面站之间的连线
   * @param group: Group 连线所在组，用于添加到场景中
   * @param satelliteLocations: {sate_id: {long: float, lat: float, height: float}}
   * @param groundPositions: {ground_id: {long: float, lat: float}}
   * @param groundConnections: {ground_id: sate_id, ...}
   */
    if (!sateLocations || !groundPositions || !groundConnections) {
      return;
    }
    const lineGroup = group.getObjectByName("lineGroup");
    for (let key in groundConnections) {
      // 判断地面站是否与卫星连接，未连接时groundConnections[key]为空
      if (groundConnections[key] === '') {
        continue;
      }
      let start_position = groundPositions[key];
      start_position['height'] = - 14 * 63782;
      let end_position = sateLocations[groundConnections[key]];
      addLine(lineGroup, start_position, end_position, '#ff0011');
    }
}

export const addLine = (
    group: Group,
    start_sate,
    end_sate,
    color: string,
) => {
  /**
   * @description: 卫星之间的连线
   * @param group: Group 连线所在组，用于添加到场景中
   * @param start_sate: {long: float, lat: float, height: float}
   * @param end_sate: {long: float, lat: float, height: float}
   * @param color: string 16进制颜色
   */
  let start_xyz = sateLocation2xyz(start_sate);
  let end_xyz = sateLocation2xyz(end_sate);
  let start_point = new Vector3(start_xyz.x, start_xyz.y, start_xyz.z);
  let end_point = new Vector3(end_xyz.x, end_xyz.y, end_xyz.z);

  // 直线
  let points = new LineCurve3(start_point, end_point).getPoints(100);
  const geometry = new BufferGeometry();
  geometry.setFromPoints(points);

  const material = new LineBasicMaterial({color: color});
  const line = new Line(geometry, material);
  group.add(line);

  // // 圆弧线
  // let curve_points = [];
  // curve_points.push(start_point);
  // const num = 500;
  // for (let i = 1; i <= num; i++) {
  //     let r = (start_sate.height * (num - i) / num + end_sate.height * i / num) / 63782 + 100;
  //     let point1 = new Vector3().copy(start_point).multiplyScalar((num - i) / num);
  //     let point2 = new Vector3().copy(end_point).multiplyScalar(i / num);
  //     curve_points.push(point1.add(point2).setLength(r));
  // }
  // curve_points.push(end_point);
  //
  // const curve = new CatmullRomCurve3(curve_points);
  // const points = curve.getPoints(1000);
  // const geometry = new BufferGeometry();
  // geometry.setFromPoints(points);
  // const material = new LineBasicMaterial({color: color});
  // const curveObject = new Line(geometry, material);
  // group.add(curveObject);

  // send(group, [curve_points]);  // 数据包展示（红色）
}

// function send(group: Group, points: Array) {
//   const aGroup = new Group();
//   for (let i = 0; i < points.length; i++) {
//     const aGeo = new SphereGeometry(1, 1, 1);
//     const aMater = new MeshPhongMaterial({color: '#ff0000'});
//     const aMesh = new Mesh(aGeo, aMater);
//     aGroup.add(aMesh);
//   }
//   let vIndex = 0;
//
//   function animateLine() {
//     aGroup.children.forEach((elem, index) => {
//       const v = points[index][vIndex];
//       elem.position.set(v.x, v.y, v.z);
//     });
//     vIndex++;
//     if (vIndex > 500) {
//       vIndex = 0;
//     }
//     setTimeout(animateLine, 10);
//   }
//
//   group.add(aGroup);
//   animateLine();
// }