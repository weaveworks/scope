import { line, curveCardinalClosed } from 'd3-shape';
import range from 'lodash/range';

const shapeSpline = line().curve(curveCardinalClosed.tension(0.65));

export function nodeShapePolygon(radius, n) {
  const innerAngle = (2 * Math.PI) / n;
  return shapeSpline(range(0, n).map(k => [
    radius * Math.sin(k * innerAngle),
    -radius * Math.cos(k * innerAngle)
  ]));
}
