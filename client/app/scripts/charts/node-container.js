import _ from 'lodash';
import React from 'react';
import { connect } from 'react-redux';
import { Motion, spring } from 'react-motion';

import { round } from '../utils/math-utils';
import Node from './node';

class NodeContainer extends React.Component {
  render() {
    const { dx, dy, focused, layoutPrecision, zoomScale } = this.props;
    const animConfig = [80, 20]; // stiffness, damping
    const scaleFactor = focused ? (1 / zoomScale) : 1;
    const other = _.omit(this.props, 'dx', 'dy');

    return (
      <Motion style={{
        x: spring(dx, animConfig),
        y: spring(dy, animConfig),
        f: spring(scaleFactor, animConfig)
      }}>
        {interpolated => {
          const transform = `translate(${round(interpolated.x, layoutPrecision)},`
            + `${round(interpolated.y, layoutPrecision)})`;
          return <Node {...other} transform={transform} scaleFactor={interpolated.f} />;
        }}
      </Motion>
    );
  }
}

export default connect()(NodeContainer);
