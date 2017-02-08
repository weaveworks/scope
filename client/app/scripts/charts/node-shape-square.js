import React from 'react';
import classNames from 'classnames';

import {
  getMetricValue,
  getMetricColor,
  getClipPathDefinition,
  renderMetricValue,
} from '../utils/metric-utils';
import {
  NODE_SHAPE_HIGHLIGHT_RADIUS,
  NODE_SHAPE_BORDER_RADIUS,
  NODE_SHAPE_SHADOW_RADIUS,
} from '../constants/styles';


export default function NodeShapeSquare({ id, highlighted, color, rx = 0, ry = 0, metric }) {
  const { height, hasMetric, formattedValue } = getMetricValue(metric);
  const metricStyle = { fill: getMetricColor(metric) };

  const className = classNames('shape', 'shape-square', { metrics: hasMetric });
  const rectProps = (scale, borderRadiusAdjustmentFactor = 1) => ({
    width: scale * 2,
    height: scale * 2,
    rx: scale * rx * borderRadiusAdjustmentFactor,
    ry: scale * ry * borderRadiusAdjustmentFactor,
    x: -scale,
    y: -scale
  });
  const clipId = `mask-${id}`;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, height)}
      {highlighted && <rect className="highlighted" {...rectProps(NODE_SHAPE_HIGHLIGHT_RADIUS)} />}
      <rect className="border" stroke={color} {...rectProps(NODE_SHAPE_BORDER_RADIUS)} />
      <rect className="shadow" {...rectProps(NODE_SHAPE_SHADOW_RADIUS, 0.85)} />
      {hasMetric && <rect
        className="metric-fill"
        clipPath={`url(#${clipId})`}
        style={metricStyle}
        {...rectProps(NODE_SHAPE_SHADOW_RADIUS, 0.85)}
      />}
      {renderMetricValue(formattedValue, highlighted && hasMetric)}
    </g>
  );
}
