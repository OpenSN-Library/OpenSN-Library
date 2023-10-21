import {
  AdditiveBlending,
  BufferGeometry,
  Color,
  Float32BufferAttribute,
  Points,
  PointsMaterial,
  Scene,
  Vector3
} from "three";

export const initBackground = (scene: Scene) => {
  const positions = [];
  const colors = [];
  const geometry = new BufferGeometry();
  for (let i = 0; i < 100; i++) {
    let vertex = new Vector3();
    vertex.x = Math.random() * 2 - 1;
    vertex.y = Math.random() * 2 - 1;
    vertex.z = Math.random() * 2 - 1;
    positions.push(vertex.x, vertex.y, vertex.z);
    let color = new Color();
    color.setHSL(Math.random() * 0.2 + 0.5, 0.55, Math.random() * 0.25 + 0.55);
    colors.push(color.r, color.g, color.b);
  }
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3));
  geometry.setAttribute('color', new Float32BufferAttribute(colors, 3));

  let starsMaterial = new PointsMaterial({
    size: 1,
    transparent: true,
    opacity: 1,
    vertexColors: true,
    blending: AdditiveBlending,
    sizeAttenuation: true
  });

  let stars = new Points(geometry, starsMaterial);
  stars.scale.set(300, 300, 300);
  scene.add(stars);
}