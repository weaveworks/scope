import React from 'react';
import { isContrastMode } from '../utils/contrast-utils';

export default function NodeShapeStack(props) {
  const contrastMode = isContrastMode();
  const Shape = props.shape;
  const [dx, dy] = contrastMode ? [0, 8] : [0, 5];
  const dsx = (props.size + dx) / props.size;
  const dsy = (props.size + dy) / props.size;
  const hls = [dsx, dsy];

  return (
    <g transform={`translate(${dx * -1}, ${dy * -2.5})`} className="stack">
      <g transform={`scale(${hls})translate(${dx}, ${dy})`} className="onlyHighlight">
        <Shape {...props} />
      </g>
      <g transform={`translate(${dx * 2}, ${dy * 2})`}>
        <Shape {...props} />
      </g>
      <g transform={`translate(${dx * 1}, ${dy * 1})`}>
        <Shape {...props} />
      </g>
      <g className="onlyMetrics">
        <Shape {...props} />
      </g>
    </g>
  );
}
