import React from 'react';
import { omit } from 'lodash';
import { Motion, spring } from 'react-motion';

import { NODES_SPRING_ANIMATION_CONFIG } from '../constants/animation';
import { isContrastMode } from '../utils/contrast-utils';
import Node from './node';


const nodeBlurOpacity = isContrastMode() ? 0.6 : 0.25;

const transformedNode = (otherProps, { x, y, k, opacity }) => (
  // NOTE: Controlling blurring and transform from here seems to re-render
  // faster than adding a CSS class and controlling it from there.
  <g transform={`translate(${x},${y}) scale(${k})`} style={{opacity}}>
    <Node {...otherProps} />
  </g>
);

export default class NodeContainer extends React.PureComponent {
  render() {
    const { dx, dy, isAnimated, scale, blurred } = this.props;
    const forwardedProps = omit(this.props, 'dx', 'dy', 'isAnimated', 'scale', 'blurred');
    const opacity = blurred ? nodeBlurOpacity : 1;

    if (!isAnimated) {
      // Show static node for optimized rendering
      return transformedNode(forwardedProps, { x: dx, y: dy, k: scale, opacity });
    }

    return (
      // Animate the node if the layout is sufficiently small
      <Motion
        style={{
          x: spring(dx, NODES_SPRING_ANIMATION_CONFIG),
          y: spring(dy, NODES_SPRING_ANIMATION_CONFIG),
          k: spring(scale, NODES_SPRING_ANIMATION_CONFIG),
          opacity: spring(opacity, NODES_SPRING_ANIMATION_CONFIG),
        }}>
        {interpolated => transformedNode(forwardedProps, interpolated)}
      </Motion>
    );
  }
}
