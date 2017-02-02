import React from 'react';
import { connect } from 'react-redux';
import { Motion, spring } from 'react-motion';
import { Map as makeMap } from 'immutable';
import { line, curveBasis } from 'd3-shape';
import { each, omit, times, constant } from 'lodash';

import { uniformSelect } from '../utils/array-utils';
import { round } from '../utils/math-utils';
import Edge from './edge';

// Spring stiffness & damping respectively
const ANIMATION_CONFIG = [80, 20];
// Tweak this value for the number of control
// points along the edge curve, e.g. values:
//   * 2 -> edges are simply straight lines
//   * 4 -> minimal value for loops to look ok
const WAYPOINTS_CAP = 8;

const spline = line()
  .curve(curveBasis)
  .x(d => d.x)
  .y(d => d.y);

const buildPath = (points, layoutPrecision) => {
  const extracted = [];
  each(points, (value, key) => {
    const axis = key[0];
    const index = key.slice(1);
    if (!extracted[index]) {
      extracted[index] = {};
    }
    extracted[index][axis] = round(value, layoutPrecision);
  });
  return extracted;
};

class EdgeContainer extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      pointsMap: makeMap()
    };
  }

  componentWillMount() {
    this.preparePoints(this.props.points);
  }

  componentWillReceiveProps(nextProps) {
    // immutablejs allows us to `===`! \o/
    if (nextProps.points !== this.props.points) {
      this.preparePoints(nextProps.points);
    }
  }

  render() {
    const { layoutPrecision, points } = this.props;
    const other = omit(this.props, 'points');

    if (layoutPrecision === 0) {
      const path = spline(points.toJS());
      return <Edge {...other} path={path} />;
    }

    return (
      <Motion style={this.state.pointsMap.toJS()}>
        {(interpolated) => {
          // convert points to path string, because that lends itself to
          // JS-equality checks in the child component
          const path = spline(buildPath(interpolated, layoutPrecision));
          return <Edge {...other} path={path} />;
        }}
      </Motion>
    );
  }

  preparePoints(nextPoints) {
    nextPoints = nextPoints.toJS();

    // Motion requires a constant number of waypoints along the path of each edge
    // for the animation to work correctly, but dagre might be changing their number
    // depending on the dynamic topology reconfiguration. Here we are transforming
    // the waypoints array given by dagre to the fixed size of `WAYPOINTS_CAP` that
    // Motion could take over.
    const pointsMissing = WAYPOINTS_CAP - nextPoints.length;
    if (pointsMissing > 0) {
      // Whenever there are some waypoints missing, we simply populate the beginning of the
      // array with the first element, as this leaves the curve interpolation unchanged.
      nextPoints = times(pointsMissing, constant(nextPoints[0])).concat(nextPoints);
    } else if (pointsMissing < 0) {
      // If there are 'too many' waypoints given by dagre, we select a sub-array of
      // uniformly distributed indices. Note that it is very important to keep the first
      // and the last endpoints in the array as they are the ones connecting the nodes.
      nextPoints = uniformSelect(nextPoints, WAYPOINTS_CAP);
    }

    let { pointsMap } = this.state;
    nextPoints.forEach((point, index) => {
      pointsMap = pointsMap.set(`x${index}`, spring(point.x, ANIMATION_CONFIG));
      pointsMap = pointsMap.set(`y${index}`, spring(point.y, ANIMATION_CONFIG));
    });

    this.setState({ pointsMap });
  }

}

export default connect()(EdgeContainer);
