import React from 'react';
import {
  NODE_SHAPE_HIGHLIGHT_RADIUS,
  NODE_SHAPE_BORDER_RADIUS,
  NODE_SHAPE_SHADOW_RADIUS,
  NODE_SHAPE_DOT_RADIUS,
} from '../constants/styles';

// This path is already normalized so no rescaling is needed.
const CLOUD_PATH = 'M-1.25 0.233Q-1.25 0.44-1.104 0.587-0.957 0.733-0.75 0.733H0.667Q0.908 '
  + '0.733 1.079 0.562 1.25 0.391 1.25 0.15 1.25-0.022 1.158-0.164 1.065-0.307 0.914-0.377q'
  + '0.003-0.036 0.003-0.056 0-0.276-0.196-0.472-0.195-0.195-0.471-0.195-0.206 0-0.373 0.115'
  + '-0.167 0.115-0.244 0.299-0.091-0.081-0.216-0.081-0.138 0-0.236 0.098-0.098 0.098-0.098 '
  + '0.236 0 0.098 0.054 0.179-0.168 0.039-0.278 0.175-0.109 0.136-0.109 0.312z';

export default function NodeShapeCloud({highlighted, color}) {
  const pathProps = r => ({ d: CLOUD_PATH, transform: `scale(${r})` });

  return (
    <g className="shape shape-cloud">
      {highlighted && <path className="highlighted" {...pathProps(NODE_SHAPE_HIGHLIGHT_RADIUS)} />}
      <path className="border" stroke={color} {...pathProps(NODE_SHAPE_BORDER_RADIUS)} />
      <path className="shadow" {...pathProps(NODE_SHAPE_SHADOW_RADIUS)} />
      <circle className="node" r={NODE_SHAPE_DOT_RADIUS} />
    </g>
  );
}
