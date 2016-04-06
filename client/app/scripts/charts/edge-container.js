import _ from 'lodash';
import d3 from 'd3';
import React from 'react';
import { Motion, spring } from 'react-motion';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';
import { Map as makeMap } from 'immutable';

import Edge from './edge';

const animConfig = [80, 20]; // stiffness, damping
const pointCount = 30;

const line = d3.svg.line()
  .interpolate('basis')
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
    extracted[index][axis] = d3.round(value, layoutPrecision);
  });
  return extracted;
};

export default class EdgeContainer extends React.Component {

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
    this.preparePoints(nextProps.points);
  }

  render() {
    const { layoutPrecision, points } = this.props;
    const other = _.omit(this.props, 'points');

    if (layoutPrecision === 0) {
      const path = line(points.toJS());
      return <Edge {...other} path={path} />;
    }

    return (
      <Motion style={this.state.pointsMap.toJS()}>
        {(interpolated) => {
          // convert points to path string, because that lends itself to
          // JS-equality checks in the child component
          const path = line(buildPath(interpolated, layoutPrecision));
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

reactMixin.onClass(EdgeContainer, PureRenderMixin);
