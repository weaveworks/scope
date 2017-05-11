import React from 'react';
import { Motion, spring } from 'react-motion';
import { Map as makeMap } from 'immutable';
import { line, curveBasis } from 'd3-shape';
import { each, omit, times } from 'lodash';

import { NODES_SPRING_ANIMATION_CONFIG } from '../constants/animation';
import { EDGE_WAYPOINTS_CAP } from '../constants/styles';
import Edge from './edge';


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
  const waypointsCount = Object.keys(waypointsMap).length / 2;
  const waypointsArray = times(waypointsCount, () => ({ x: 0, y: 0 }));
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
    while (nextWaypoints.size < EDGE_WAYPOINTS_CAP) {
      nextWaypoints = nextWaypoints.insert(0, nextWaypoints.first());
    }

    let { waypointsMap } = this.state;
    nextWaypoints.toJS().forEach((point, index) => {
      waypointsMap = waypointsMap.set(`x${index}`, spring(point.x, NODES_SPRING_ANIMATION_CONFIG));
      waypointsMap = waypointsMap.set(`y${index}`, spring(point.y, NODES_SPRING_ANIMATION_CONFIG));
    });

    this.setState({ waypointsMap });
  }
}
