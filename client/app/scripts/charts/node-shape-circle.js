import React from 'react';
import classNames from 'classnames';
import {getMetricValue, getMetricColor} from '../utils/metric-utils.js';
import {CANVAS_METRIC_FONT_SIZE} from '../constants/styles.js';

export default function NodeShapeCircle({id, highlighted, size, color, metric}) {
  const hightlightNode = <circle r={size * 0.7} className="highlighted" />;
  const clipId = `mask-${id}`;
  const {height, value, formattedValue} = getMetricValue(metric, size);

  const className = classNames('shape', {
    metrics: value !== null
  });
  const metricStyle = {
    fill: getMetricColor(metric)
  };
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

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
