import React from 'react';
import range from 'lodash/range';
import { line, curveCardinalClosed } from 'd3-shape';

import { UNIT_CLOUD_PATH, UNIT_CYLINDER_PATH } from '../constants/styles';


export const pathElement = React.createFactory('path');
export const circleElement = React.createFactory('circle');
export const rectangleElement = React.createFactory('rect');

function curvedUnitPolygonPath(n) {
  const curve = curveCardinalClosed.tension(0.65);
  const spline = line().curve(curve);
  const innerAngle = (2 * Math.PI) / n;
  return spline(range(0, n).map(k => [
    Math.sin(k * innerAngle),
    -Math.cos(k * innerAngle),
  ]));
}

export const circleShapeProps = { r: 1 };
export const triangleShapeProps = { d: curvedUnitPolygonPath(3) };
export const squareShapeProps = {
  width: 1.8, height: 1.8, rx: 0.4, ry: 0.4, x: -0.9, y: -0.9
};
export const pentagonShapeProps = { d: curvedUnitPolygonPath(5) };
export const hexagonShapeProps = { d: curvedUnitPolygonPath(6) };
export const heptagonShapeProps = { d: curvedUnitPolygonPath(7) };
export const octagonShapeProps = { d: curvedUnitPolygonPath(8) };
export const cloudShapeProps = { d: UNIT_CLOUD_PATH };
export const cylinderShapeProps = { d: UNIT_CYLINDER_PATH };
export const dottedCylinderShapeProps = { d: UNIT_CYLINDER_PATH };
