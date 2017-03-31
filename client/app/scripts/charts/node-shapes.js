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
  cloudShapeProps,
  circleShapeProps,
  squareShapeProps,
  hexagonShapeProps,
  heptagonShapeProps,
} from '../utils/node-shape-utils';


function NodeShape(shapeType, renderShapeLayer, shapeProps, { id, highlighted, color, metric }) {
  const { height, hasMetric, formattedValue } = getMetricValue(metric);
  const className = classNames('shape', `shape-${shapeType}`, { metrics: hasMetric });
  const metricStyle = { fill: getMetricColor(metric) };
  const clipId = `mask-${id}`;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, height)}
      {highlighted && renderShapeLayer({
        className: 'highlighted',
        transform: `scale(${NODE_BASE_SIZE * 0.7})`,
        ...shapeProps,
      })}
      {renderShapeLayer({
        className: 'border',
        transform: `scale(${NODE_BASE_SIZE * 0.5})`,
        stroke: color,
        ...shapeProps,
      })}
      {renderShapeLayer({
        className: 'shadow',
        transform: `scale(${NODE_BASE_SIZE * 0.45})`,
        ...shapeProps,
      })}
      {hasMetric && renderShapeLayer({
        className: 'metric-fill',
        transform: `scale(${NODE_BASE_SIZE * 0.45})`,
        clipPath: `url(#${clipId})`,
        style: metricStyle,
        ...shapeProps,
      })}
      {hasMetric && highlighted ?
        <text>{formattedValue}</text> :
        <circle className="node" r={NODE_BASE_SIZE * 0.1} />
      }
    </g>
  );
}

export const NodeShapeCloud = props => NodeShape('cloud', pathElement, cloudShapeProps, props);
export const NodeShapeCircle = props => NodeShape('circle', circleElement, circleShapeProps, props);
export const NodeShapeHexagon = props => NodeShape('hexagon', pathElement, hexagonShapeProps, props);
export const NodeShapeHeptagon = props => NodeShape('heptagon', pathElement, heptagonShapeProps, props);
export const NodeShapeSquare = props => NodeShape('square', rectangleElement, squareShapeProps, props);
