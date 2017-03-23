import React from 'react';
import { Motion, spring } from 'react-motion';
import { Map as makeMap } from 'immutable';
import { line, curveBasis } from 'd3-shape';
import { each, omit, times, constant } from 'lodash';

import { NODES_SPRING_ANIMATION_CONFIG } from '../constants/animation';
import { uniformSelect } from '../utils/array-utils';
import Edge from './edge';

// Tweak this value for the number of control
// points along the edge curve, e.g. values:
//   * 2 -> edges are simply straight lines
//   * 4 -> minimal value for loops to look ok
const WAYPOINTS_COUNT = 8;

const spline = line()
  .curve(curveBasis)
  .x(d => d.x)
  .y(d => d.y);

const transformedEdge = (props, path) => (
  <Edge {...props} path={spline(path)} />
);

// Converts a waypoints map of the format {x0: 11, y0: 22, x1: 33, y1: 44}
// that is used by Motion to an array of waypoints in the format
// [{x: 11, y: 22}, {x: 33, y: 44}] that can be used by D3.
const waypointsMapToArray = (waypointsMap) => {
  const waypointsArray = times(WAYPOINTS_COUNT, () => ({ x: 0, y: 0}));
  each(waypointsMap, (value, key) => {
    const [axis, index] = [key[0], key.slice(1)];
    waypointsArray[index][axis] = value;
  });
  return waypointsArray;
};


export default class EdgeContainer extends React.PureComponent {
  constructor(props, context) {
    super(props, context);
    this.state = { waypointsMap: makeMap() };
  }

  componentWillMount() {
    if (this.props.isAnimated) {
      this.prepareWaypointsForMotion(this.props.waypoints);
    }
  }

  componentWillReceiveProps(nextProps) {
    // immutablejs allows us to `===`! \o/
    if (this.props.isAnimated && nextProps.waypoints !== this.props.waypoints) {
      this.prepareWaypointsForMotion(nextProps.waypoints);
    }
  }

  render() {
    const { isAnimated, waypoints } = this.props;
    const forwardedProps = omit(this.props, 'isAnimated', 'waypoints');

    if (!isAnimated) {
      return transformedEdge(forwardedProps, waypoints.toJS());
    }

    return (
      // For the Motion interpolation to work, the waypoints need to be in a map format like
      // {x0: 11, y0: 22, x1: 33, y1: 44} that we convert to the array format when rendering.
      <Motion style={this.state.waypointsMap.toJS()}>
        {interpolated => transformedEdge(forwardedProps, waypointsMapToArray(interpolated))}
      </Motion>
    );
  }

  prepareWaypointsForMotion(nextWaypoints) {
    nextWaypoints = nextWaypoints.toJS();

    // Motion requires a constant number of waypoints along the path of each edge
    // for the animation to work correctly, but dagre might be changing their number
    // depending on the dynamic topology reconfiguration. Here we are transforming
    // the waypoints array given by dagre to the fixed size of `WAYPOINTS_COUNT` that
    // Motion could take over.
    const waypointsMissing = WAYPOINTS_COUNT - nextWaypoints.length;
    if (waypointsMissing > 0) {
      // Whenever there are some waypoints missing, we simply populate the beginning of the
      // array with the first element, as this leaves the curve interpolation unchanged.
      nextWaypoints = times(waypointsMissing, constant(nextWaypoints[0])).concat(nextWaypoints);
    } else if (waypointsMissing < 0) {
      // If there are 'too many' waypoints given by dagre, we select a sub-array of
      // uniformly distributed indices. Note that it is very important to keep the first
      // and the last endpoints in the array as they are the ones connecting the nodes.
      nextWaypoints = uniformSelect(nextWaypoints, WAYPOINTS_COUNT);
    }

    let { waypointsMap } = this.state;
    nextWaypoints.forEach((point, index) => {
      waypointsMap = waypointsMap.set(`x${index}`, spring(point.x, NODES_SPRING_ANIMATION_CONFIG));
      waypointsMap = waypointsMap.set(`y${index}`, spring(point.y, NODES_SPRING_ANIMATION_CONFIG));
    });

    this.setState({ waypointsMap });
  }
}
