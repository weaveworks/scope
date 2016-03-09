import React from 'react';
import d3 from 'd3';
import {getMetricValue} from '../utils/data-utils.js';

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


export default function NodeShapeHex({highlighted, size, color, metric}) {
  const pathProps = v => ({
    d: getPoints(size * v * 2),
    transform: `rotate(90) translate(-${size * getWidth(v)}, -${size * v})`
  });

  const shadowSize = 0.45;
  const upperHexBitHeight = -0.25 * size * shadowSize;

  const clipId = `mask-${Math.random()}`;
  const {height, value, formattedValue} = getMetricValue(metric, size);

  return (
    <g className="shape">
      <defs>
        <clipPath id={clipId}>
          <rect
            width={size}
            height={size}
            x={size - height + upperHexBitHeight}
            />
        </clipPath>
      </defs>
      {highlighted && <path className="highlighted" {...pathProps(0.7)} />}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(shadowSize)} />
      <path className="metric-fill" clipPath={`url(#${clipId})`} {...pathProps(shadowSize)} />
      {highlighted && value !== null ?
        <text dy="0.35em" style={{'textAnchor': 'middle'}}>
          {formattedValue}
        </text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}
