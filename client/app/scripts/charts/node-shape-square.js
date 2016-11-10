import React from 'react';
import classNames from 'classnames';
import {getMetricValue, getMetricColor, getClipPathDefinition} from '../utils/metric-utils.js';
import {CANVAS_METRIC_FONT_SIZE} from '../constants/styles.js';


export default function NodeShapeSquare({
  id, highlighted, size, rx = 0, ry = 0, metric, color, lightColor, darkColor
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
      <rect className="outline" stroke={darkColor} fill="none" {...rectProps(0.55, 0.55)} />
      <rect className="border" stroke={color} fill={lightColor} {...rectProps(0.5, 0.5)} />
      {hasMetric && <rect className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`} {...rectProps(0.45, 0.39)} />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>
          {formattedValue}
        </text> :
        <circle className="node" r={Math.max(1.33333, (size * 0.08333))} fill={darkColor} />}
    </g>
  );
}
