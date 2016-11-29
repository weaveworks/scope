import _ from 'lodash';
import React from 'react';
import { connect } from 'react-redux';
import { Motion, spring } from 'react-motion';
import { Map as makeMap } from 'immutable';
import { line, curveBasis } from 'd3-shape';

import { round } from '../utils/math-utils';
import Edge from './edge';

const animConfig = [80, 20]; // stiffness, damping
const pointCount = 30;

const spline = line()
  .curve(curveBasis)
  .x(d => d.x)
  .y(d => d.y);

const buildPath = (points, layoutPrecision) => {
  const extracted = [];
  _.each(points, (value, key) => {
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
    const other = _.omit(this.props, 'points');

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
    // Spring needs constant field count, hoping that dagre will insert never more than `pointCount`
    let { pointsMap } = this.state;

    // filling up the map with copies of the first point
    const filler = nextPoints.first();
    const missing = pointCount - nextPoints.size;
    let index = 0;
    if (missing > 0) {
      while (index < missing) {
        pointsMap = pointsMap.set(`x${index}`, spring(filler.get('x'), animConfig));
        pointsMap = pointsMap.set(`y${index}`, spring(filler.get('y'), animConfig));
        index++;
      }
    }

    nextPoints.forEach((point, i) => {
      pointsMap = pointsMap.set(`x${index + i}`, spring(point.get('x'), animConfig));
      pointsMap = pointsMap.set(`y${index + i}`, spring(point.get('y'), animConfig));
    });

    this.setState({ pointsMap });
  }

}

export default connect()(EdgeContainer);
