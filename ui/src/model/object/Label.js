import {CSS3DSprite} from "three/examples/jsm/renderers/CSS3DRenderer";

export const Label = (text: string, xyz) => {
  /**
   * @description: 为卫星创建标签
   * @param text: String 标签内容
   * @param xyz: 卫星位置
   */
  let labelDiv = document.createElement('div');
  labelDiv.className = 'label';
  labelDiv.textContent = text;

  let pointLabel = new CSS3DSprite(labelDiv);
  pointLabel.position.set(xyz.x, xyz.y, xyz.z);
  let scale = 0.25;
  pointLabel.scale.set(scale, scale, scale);
  pointLabel.name = text;

  return pointLabel;
}