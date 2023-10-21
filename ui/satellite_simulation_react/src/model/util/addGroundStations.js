import {Group} from 'three';
import {GroundStation} from "../object/GroundStation";

export const addGroundStations = (groundPositions, group) => {
  /**
   * @description 根据地面站的位置信息，添加地面站
   * @param groundPositions {Array} 地面站的位置信息
   * @param group {Group} 场景组
   */
  const groundGroup = new Group();

  for (let key in groundPositions) {
    const groundPosition = groundPositions[key];
    const groundStation = GroundStation(groundPosition, key);
    groundStation.name = key;
    groundGroup.add(groundStation);
  }

  groundGroup.name = 'groundGroup';
  group.add(groundGroup);
};