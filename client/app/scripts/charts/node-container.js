import React from 'react';
import { omit } from 'lodash';
import { connect } from 'react-redux';
import { Motion, spring } from 'react-motion';

import { NODES_SPRING_ANIMATION_CONFIG } from '../constants/animation';
import { NODE_BLUR_OPACITY } from '../constants/styles';
import Node from './node';

const transformedNode = (otherProps, { x, y, k }) => (
  <Node transform={`translate(${x},${y}) scale(${k})`} {...otherProps} />
);

class NodeContainer extends React.Component {
  render() {
    const { dx, dy, isAnimated, scale, blurred } = this.props;
    const forwardedProps = omit(this.props, 'dx', 'dy', 'isAnimated', 'scale', 'blurred');
    const opacity = blurred ? NODE_BLUR_OPACITY : 1;

    // NOTE: Controlling blurring from here seems to re-render faster
    // than adding a CSS class and controlling it from there.
    return (
      <g className="node-container" style={{opacity}}>
        {!isAnimated ?

        // Show static node for optimized rendering
        transformedNode(forwardedProps, { x: dx, y: dy, k: scale }) :

        // Animate the node if the layout is sufficiently small
        <Motion
          style={{
            x: spring(dx, NODES_SPRING_ANIMATION_CONFIG),
            y: spring(dy, NODES_SPRING_ANIMATION_CONFIG),
            k: spring(scale, NODES_SPRING_ANIMATION_CONFIG)
          }}>
          {interpolated => transformedNode(forwardedProps, interpolated)}
        </Motion>}
      </g>
    );
  }
}

export default connect()(NodeContainer);
