import React from 'react';
import { Motion } from 'react-motion';
import { Repeat, fromJS, Map as makeMap } from 'immutable';
import { line, curveBasis } from 'd3-shape';
import { times } from 'lodash';

import { weakSpring } from 'weaveworks-ui-components/lib/utils/animation';

import { NODE_BASE_SIZE, EDGE_WAYPOINTS_CAP } from '../constants/styles';
import Edge from './edge';


const spline = line()
  .curve(curveBasis)
  .x(d => d.x)
  .y(d => d.y);

const transformedEdge = (props, path, thickness) => (
  <Edge {...props} path={spline(path)} thickness={thickness} />
);

// Converts a waypoints map of the format { x0: 11, y0: 22, x1: 33, y1: 44 }
// that is used by Motion to an array of waypoints in the format
// [{ x: 11, y: 22 }, { x: 33, y: 44 }] that can be used by D3.
const waypointsMapToArray = (waypointsMap) => {
  const waypointsArray = times(EDGE_WAYPOINTS_CAP, () => ({}));
  waypointsMap.forEach((value, key) => {
    const [axis, index] = [key[0], key.slice(1)];
    waypointsArray[index][axis] = value;
  });
  return waypointsArray;
};

// Converts a waypoints array of the input format [{ x: 11, y: 22 }, { x: 33, y: 44 }]
// to an array of waypoints that is used by Motion in the format { x0: 11, y0: 22, x1: 33, y1: 44 }.
const waypointsArrayToMap = (waypointsArray) => {
  let waypointsMap = makeMap();
  waypointsArray.forEach((point, index) => {
    waypointsMap = waypointsMap.set(`x${index}`, weakSpring(point.get('x')));
    waypointsMap = waypointsMap.set(`y${index}`, weakSpring(point.get('y')));
  });
  return waypointsMap;
};


export default class EdgeContainer extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = {
      thickness: 1,
      waypointsMap: makeMap(),
    };
  }

  componentWillMount() {
    this.prepareWaypointsForMotion(this.props);
  }

  componentWillReceiveProps(nextProps) {
    // immutablejs allows us to `===`! \o/
    const waypointsChanged = this.props.waypoints !== nextProps.waypoints;
    const animationChanged = this.props.isAnimated !== nextProps.isAnimated;
    if (waypointsChanged || animationChanged) {
      this.prepareWaypointsForMotion(nextProps);
    }
    // Edge thickness will reflect the zoom scale.
    const baseScale = (nextProps.scale * 0.01) * NODE_BASE_SIZE;
    const thickness = (nextProps.focused ? 3 : 1) * baseScale;
    this.setState({ thickness });
  }

  render() {
    const {
      isAnimated, waypoints, scale, ...forwardedProps
    } = this.props;
    const { thickness, waypointsMap } = this.state;

    if (!isAnimated) {
      return transformedEdge(forwardedProps, waypoints.toJS(), thickness);
    }

    return (
      // For the Motion interpolation to work, the waypoints need to be in a map format like
      // { x0: 11, y0: 22, x1: 33, y1: 44 } that we convert to the array format when rendering.
      <Motion style={{
        interpolatedThickness: weakSpring(thickness),
        ...waypointsMap.toJS(),
      }}>
        {
          ({ interpolatedThickness, ...interpolatedWaypoints }) =>
            transformedEdge(
              forwardedProps,
              waypointsMapToArray(fromJS(interpolatedWaypoints)),
              interpolatedThickness
            )
        }
      </Motion>
    );
  }

  prepareWaypointsForMotion({ waypoints, isAnimated }) {
    // Don't update if the edges are not animated.
    if (!isAnimated) return;

    // The Motion library requires the number of waypoints to be constant, so we fill in for
    // the missing ones by reusing the edge source point, which doesn't affect the edge shape
    // because of how the curveBasis interpolation is done.
    const waypointsMissing = EDGE_WAYPOINTS_CAP - waypoints.size;
    if (waypointsMissing > 0) {
      waypoints = Repeat(waypoints.get(0), waypointsMissing).concat(waypoints);
    }

    this.setState({ waypointsMap: waypointsArrayToMap(waypoints) });
  }
}
