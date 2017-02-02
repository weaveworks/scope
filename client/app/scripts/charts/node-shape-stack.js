import React from 'react';
import { isContrastMode } from '../utils/contrast-utils';

export default function NodeShapeStack(props) {
  const dy = isContrastMode() ? 0.15 : 0.1;
  const highlightScale = [1, 1 + dy];

  const Shape = props.shape;
  return (
    <g transform={`translate(0, ${dy * -2.5})`} className="stack">
      <g transform={`scale(${highlightScale}) translate(0, ${dy})`} className="highlight">
        <Shape {...props} />
      </g>
      <g transform={`translate(0, ${dy * 2})`}>
        <Shape {...props} />
      </g>
      <g transform={`translate(0, ${dy * 1})`}>
        <Shape {...props} />
      </g>
      <g className="only-metrics">
        <Shape {...props} />
      </g>
    </g>
  );
}
