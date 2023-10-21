import {Group} from "three";
import {addLine} from "./addLines";

export const addRoutePath = (group, routePath, sateLocations) => {
  /**
   * @description: 添加路由路径
   * @param group: Group
   * @param routePath: Array
   * @param @param sateLocations: {sate_id: {long: number, lat: number, height: number}}
   */
  if (routePath.length === 0) {
    return;
  }
  let pathGroup = new Group();
  let start;
  let end = routePath[0]["node_id"];
  for (let i = 1; i < routePath.length; i++) {
    start = end;
    end = routePath[i]["node_id"];
    addLine(pathGroup, sateLocations[start], sateLocations[end], '#ff0000')
  }
  pathGroup.name = "pathGroup";
  this.state.group.add(pathGroup);
}