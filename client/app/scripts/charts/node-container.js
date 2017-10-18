import React from 'react';
import { Motion } from 'react-motion';

import { weakSpring } from '../utils/animation-utils';
import Node from './node';


const transformedNode = (otherProps, { x, y, k }) => (
  <g transform={`translate(${x},${y}) scale(${k})`}>
    <Node {...otherProps} />
  </g>
);

export default class NodeContainer extends React.PureComponent {
  render() {
    const {
      dx, dy, isAnimated, scale, ...forwardedProps
    } = this.props;

    if (!isAnimated) {
      // Show static node for optimized rendering
      return transformedNode(forwardedProps, { x: dx, y: dy, k: scale });
    }

    return (
      // Animate the node if the layout is sufficiently small
      <Motion style={{ x: weakSpring(dx), y: weakSpring(dy), k: weakSpring(scale) }}>
        {interpolated => transformedNode(forwardedProps, interpolated)}
      </Motion>
    );
  }
}
