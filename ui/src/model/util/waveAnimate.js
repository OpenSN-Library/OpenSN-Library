export const waveAnimate = (groundGroup) => {
  /**
   * @description: 地面站波浪动画
   * @param {groundGroup} 地面站组
   */
  groundGroup.children.forEach((ground) => {
    const mesh = ground.getObjectByName('wave');
    mesh._s += 0.007;
    const scale = mesh.size * mesh._s;
    mesh.scale.set(scale, scale, scale);
    if (mesh._s <= 1.5) {
      mesh.material.opacity = (mesh._s - 1) * 2;
    } else if (mesh._s > 1.5 && mesh._s <= 2) {
      mesh.material.opacity = 1 - (mesh._s - 1.5) * 2;
    } else {
      mesh._s = 1;
    }
  });
}