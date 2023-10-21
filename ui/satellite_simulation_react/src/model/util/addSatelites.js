import {Group} from 'three';
import {Satellite} from "../object/Satellite";
import {Label} from "../object/Label";
import {sateLocation2xyz} from "./sateLocation2xyz";

export const addSatellites = (sateLocations, group) => {
  /**
   * @description: 根据卫星位置添加卫星模型
   * @param {Array} sateLocations
   * @param {Group} group 场景组
   */
  if (!sateLocations) {
    return;
  }
  const sateGroup = new Group();
  const labelGroup = new Group();

  for (let key in sateLocations) {
    const sateLocation = sateLocations[key];

    const sate = Satellite(sateLocation, key);
    sate.name = key;
    sateGroup.add(sate);

    // 添加卫星标签
    const xyz = sateLocation2xyz(sateLocation);
    const label = Label(key, xyz);
    label.name = key;
    labelGroup.add(label);
  }

  sateGroup.name = 'sateGroup';
  labelGroup.name = 'labelGroup';
  group.add(sateGroup);
  group.add(labelGroup);
};