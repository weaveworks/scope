import React from 'react';
import classNames from 'classnames';
import {getMetricValue, getMetricColor} from '../utils/data-utils.js';

export default function NodeShapeSquare({
  highlighted, size, color, rx = 0, ry = 0, metric
}) {
  const rectProps = (v, vr) => ({
    width: v * size * 2,
    height: v * size * 2,
    rx: (vr || v) * size * rx,
    ry: (vr || v) * size * ry,
    x: -size * v,
    y: -size * v
  });

  const clipId = `mask-${Math.random()}`;
  const {height, value, formattedValue} = getMetricValue(metric, size);
  const className = classNames('shape', {
    metrics: value !== null
  });

  const metricStyle = {
    fillOpacity: 0.5,
    fill: getMetricColor()
  };

  return (
    <g className={className}>
      <defs>
        <clipPath id={clipId}>
          <rect
            width={size}
            height={size}
            x={-size * 0.5}
            y={size * 0.5 - height}
            />
        </clipPath>
      </defs>
      {highlighted && <rect className="highlighted" {...rectProps(0.7)} />}
      <rect className="border" stroke={color} {...rectProps(0.5, 0.5)} />
      <rect className="shadow" {...rectProps(0.45, 0.39)} />
      <rect className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`} {...rectProps(0.45, 0.39)} />
      {highlighted && value !== null ?
        <text dy="0.35em" style={{textAnchor: 'middle'}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}
