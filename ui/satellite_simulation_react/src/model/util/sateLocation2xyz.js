export const sateLocation2xyz = (sateLocation) => {
  /**
   * @description: 卫星位置转空间坐标
   * @param sateLocation: {lat: float, long: float, height: float}
   * @return: {x: float, y: float, z: float}
   */
  const rad = sateLocation.height / 63782 + 100;
  const theta = Math.PI / 2 - sateLocation.lat;
  const phi = - sateLocation.long;
  return {
    x: rad * Math.sin(theta) * Math.cos(phi),
    y: rad * Math.cos(theta),
    z: rad * Math.sin(theta) * Math.sin(phi)
  }
}