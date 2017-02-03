import React from 'react';
import {
  NODE_SHAPE_HIGHLIGHT_RADIUS,
  NODE_SHAPE_BORDER_RADIUS,
  NODE_SHAPE_SHADOW_RADIUS,
  NODE_SHAPE_DOT_RADIUS,
  NODE_BASE_SIZE,
} from '../constants/styles';

// This path is already normalized so no dynamic rescaling is needed.
const CLOUD_PATH = 'M-125 23.333Q-125 44.036-110.352 58.685-95.703 73.333-75 73.333H66.667Q90.755 '
  + '73.333 107.878 56.211 125 39.089 125 15 125-2.188 115.755-16.445 106.51-30.703 91.406-37.734q'
  + '0.26-3.646 0.261-5.599 0-27.604-19.532-47.136-19.531-19.531-47.135-19.531-20.573 0-37.305 '
  + '11.458-16.732 11.458-24.414 29.948-9.115-8.073-21.614-8.073-13.802 0-23.568 9.766-9.766 9.766-'
  + '9.766 23.568 0 9.766 5.339 17.968-16.797 3.906-27.735 17.513-10.938 13.607-10.937 31.185z';

export default function NodeShapeCloud({ highlighted, color }) {
  const pathProps = r => ({ d: CLOUD_PATH, transform: `scale(${r / NODE_BASE_SIZE})` });

  return (
    <g className="shape shape-cloud">
      {highlighted && <path className="highlighted" {...pathProps(NODE_SHAPE_HIGHLIGHT_RADIUS)} />}
      <path className="border" stroke={color} {...pathProps(NODE_SHAPE_BORDER_RADIUS)} />
      <path className="shadow" {...pathProps(NODE_SHAPE_SHADOW_RADIUS)} />
      <circle className="node" r={NODE_SHAPE_DOT_RADIUS} />
    </g>
  );
}
