import React from 'react';
import classNames from 'classnames';

import { getMetricValue, getMetricColor, getClipPathDefinition } from '../utils/metric-utils';
import { CANVAS_METRIC_FONT_SIZE } from '../constants/styles';
import { getBorderProps } from './node';


export default function NodeShapeSquare({
  id, highlighted, size, color, rx = 0, ry = 0, metric
}) {
  const rectProps = (scale, radiusScale) => ({
    width: scale * size * 2,
    height: scale * size * 2,
    rx: (radiusScale || scale) * size * rx,
    ry: (radiusScale || scale) * size * ry,
    x: -size * scale,
    y: -size * scale
  });

  const clipId = `mask-${id}`;
  const {height, hasMetric, formattedValue} = getMetricValue(metric, size);
  const metricStyle = { fill: getMetricColor(metric) };
  const className = classNames('shape', { metrics: hasMetric });
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, size, height)}
      {highlighted && <rect className="highlighted" {...rectProps(0.7)} />}
      <rect className="border" stroke={color} {...rectProps(0.5, 0.5)} {...getBorderProps()} />
      <rect className="shadow" {...rectProps(0.45, 0.39)} />
      {hasMetric && <rect
        className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`}
        {...rectProps(0.45, 0.39)}
      />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>
          {formattedValue}
        </text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}
