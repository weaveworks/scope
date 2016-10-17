import React from 'react';
import d3 from 'd3';
import classNames from 'classnames';
import {getMetricValue, getMetricColor, getClipPathDefinition} from '../utils/metric-utils.js';
import {CANVAS_METRIC_FONT_SIZE} from '../constants/styles.js';


const line = d3.svg.line()
  .interpolate('cardinal-closed')
  .tension(0.25);


function getWidth(h) {
  return (Math.sqrt(3) / 2) * h;
}


function getPoints(h) {
  const w = getWidth(h);
  const points = [
    [w * 0.5, 0],
    [w, 0.25 * h],
    [w, 0.75 * h],
    [w * 0.5, h],
    [0, 0.75 * h],
    [0, 0.25 * h]
  ];

  return line(points);
}


export default function NodeShapeHex({id, highlighted, size, color, metric, wireframe}) {
  const pathProps = v => ({
    d: getPoints(size * v * 2),
    transform: `rotate(90) translate(-${size * getWidth(v)}, -${size * v})`
  });

  const shadowSize = 0.45;
  const upperHexBitHeight = -0.25 * size * shadowSize;

  const clipId = `mask-${id}`;
  const {height, hasMetric, formattedValue} = getMetricValue(metric, size);
  const metricStyle = { fill: getMetricColor(metric) };
  const className = classNames('shape', { metrics: hasMetric, wireframe });
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, size, height, size - height +
                                          upperHexBitHeight, 0)}
      {highlighted && <path className="highlighted" {...pathProps(0.7)} />}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(shadowSize)} />
      {hasMetric && <path className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`} {...pathProps(shadowSize)} />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>
          {formattedValue}
        </text> :
        <circle className="node" fill={color} stroke={color} r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}
