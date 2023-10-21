import {BufferGeometry, EllipseCurve, Line, LineBasicMaterial} from "three";

export const addLongLine = (position) => {
  /**
   * @description: 绘制卫星经线
   * @param position: {lat: number, long: number, height: number} 卫星位置
   */
  let radius = position.height / 63782 + 100;
  let curve = new EllipseCurve(0, 0, radius, radius, 0, 2 * Math.PI, 0);

  let points = curve.getPoints(500);
  let geometry = new BufferGeometry().setFromPoints(points);
  let material = new LineBasicMaterial({color: 0x333d3d});

  let line = new Line(geometry, material);
  const phi = - position.long;
  line.rotateY(-phi);

  line.name = "longLine";
  return line;
}
