import React from 'react';
import classNames from 'classnames';
import {getMetricValue, getMetricColor} from '../utils/data-utils.js';

export default function NodeShapeCircle({highlighted, size, color, metric}) {
  const hightlightNode = <circle r={size * 0.7} className="highlighted" />;
  const clipId = `mask-${Math.random()}`;
  const {height, value, formattedValue} = getMetricValue(metric, size);

  const className = classNames('shape', {
    metrics: value !== null
  });
  const metricStyle = {
    fillOpacity: 0.5,
    fill: getMetricColor()
  };
  const fontSize = size * 0.19;

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
      {highlighted && hightlightNode}
      <circle r={size * 0.5} className="border" stroke={color} />
      <circle r={size * 0.45} className="shadow" />
      <circle r={size * 0.45} className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`} />
      {highlighted && value !== null ?
        <text style={{fontSize}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}
