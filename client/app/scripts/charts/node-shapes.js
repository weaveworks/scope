import React from 'react';
import classNames from 'classnames';

import { NODE_BASE_SIZE } from '../constants/styles';
import {
  getMetricValue,
  getMetricColor,
  getClipPathDefinition,
} from '../utils/metric-utils';
import {
  pathElement,
  circleElement,
  rectangleElement,
  circleShapeProps,
  triangleShapeProps,
  squareShapeProps,
  pentagonShapeProps,
  hexagonShapeProps,
  heptagonShapeProps,
  octagonShapeProps,
  cloudShapeProps,
} from '../utils/node-shape-utils';
import { encodeIdAttribute } from '../utils/dom-utils';


function NodeShape(shapeType, shapeElement, shapeProps, { id, highlighted, color, metric }) {
  const { height, hasMetric, formattedValue } = getMetricValue(metric);
  const className = classNames('shape', `shape-${shapeType}`, { metrics: hasMetric });
  const metricStyle = { fill: getMetricColor(metric) };
  const clipId = encodeIdAttribute(`metric-clip-${id}`);

  return (
    <g className={className}>
      {highlighted && shapeElement({
        className: 'highlight-border',
        transform: `scale(${NODE_BASE_SIZE * 0.5})`,
        ...shapeProps,
      })}
      {highlighted && shapeElement({
        className: 'highlight-shadow',
        transform: `scale(${NODE_BASE_SIZE * 0.5})`,
        ...shapeProps,
      })}
      {shapeElement({
        className: 'background',
        transform: `scale(${NODE_BASE_SIZE * 0.48})`,
        ...shapeProps,
      })}
      {hasMetric && getClipPathDefinition(clipId, height, 0.48)}
      {hasMetric && shapeElement({
        className: 'metric-fill',
        transform: `scale(${NODE_BASE_SIZE * 0.48})`,
        clipPath: `url(#${clipId})`,
        style: metricStyle,
        ...shapeProps,
      })}
      {shapeElement({
        className: 'shadow',
        transform: `scale(${NODE_BASE_SIZE * 0.49})`,
        ...shapeProps,
      })}
      {shapeElement({
        className: 'border',
        transform: `scale(${NODE_BASE_SIZE * 0.5})`,
        stroke: color,
        ...shapeProps,
      })}
      {hasMetric && highlighted ?
        <text>{formattedValue}</text> :
        <circle className="node" r={NODE_BASE_SIZE * 0.1} />
      }
    </g>
  );
}

export const NodeShapeCircle = props => NodeShape('circle', circleElement, circleShapeProps, props);
export const NodeShapeTriangle = props => NodeShape('triangle', pathElement, triangleShapeProps, props);
export const NodeShapeSquare = props => NodeShape('square', rectangleElement, squareShapeProps, props);
export const NodeShapePentagon = props => NodeShape('pentagon', pathElement, pentagonShapeProps, props);
export const NodeShapeHexagon = props => NodeShape('hexagon', pathElement, hexagonShapeProps, props);
export const NodeShapeHeptagon = props => NodeShape('heptagon', pathElement, heptagonShapeProps, props);
export const NodeShapeOctagon = props => NodeShape('octagon', pathElement, octagonShapeProps, props);
export const NodeShapeCloud = props => NodeShape('cloud', pathElement, cloudShapeProps, props);
