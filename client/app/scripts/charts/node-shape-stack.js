import React from 'react';

import { NODE_BASE_SIZE } from '../constants/styles';

export default function NodeShapeStack(props) {
  const dx = NODE_BASE_SIZE * (props.contrastMode ? 0.06 : 0.05);
  const dy = NODE_BASE_SIZE * (props.contrastMode ? 0.12 : 0.1);
  const translateAlongAxis = t => `translate(${t * dx}, ${t * dy})`;
  const Shape = props.shape;

  return (
    <g transform={translateAlongAxis(-2.5)} className="stack">
      <g transform={translateAlongAxis(2)}><Shape {...props} /></g>
      <g transform={translateAlongAxis(1)}><Shape {...props} highlighted={false} /></g>
      <g transform={translateAlongAxis(0)}><Shape {...props} highlighted={false} /></g>
    </g>
  );
}
