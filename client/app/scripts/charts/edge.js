const _ = require('lodash');
const d3 = require('d3');
const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');

const line = d3.svg.line()
  .interpolate('basis')
  .x(function(d) { return d.x; })
  .y(function(d) { return d.y; });

const animConfig = [80, 20]; // stiffness, bounce

const flattenPoints = function(points) {
  const flattened = {};
  points.forEach(function(point, i) {
    flattened['x' + i] = {val: point.x, config: animConfig};
    flattened['y' + i] = {val: point.y, config: animConfig};
  });
  return flattened;
};

const extractPoints = function(points) {
  const extracted = [];
  _.each(points, function(value, key) {
    const axis = key[0];
    const index = key.slice(1);
    if (!extracted[index]) {
      extracted[index] = {};
    }
    extracted[index][axis] = value.val;
  });
  return extracted;
};

const Edge = React.createClass({

  getInitialState: function() {
    return {
      points: []
    };
  },

  componentWillMount: function() {
    this.ensureSameLength(this.props.points);
  },

  componentWillReceiveProps: function(nextProps) {
    this.ensureSameLength(nextProps.points);
  },

  render: function() {
    const className = this.props.highlighted ? 'edge highlighted' : 'edge';
    const points = flattenPoints(this.props.points);
    const props = this.props;
    const handleMouseEnter = this.handleMouseEnter;
    const handleMouseLeave = this.handleMouseLeave;

    return (
      <Spring endValue={points}>
        {function(interpolated) {
          const path = line(extractPoints(interpolated));
          return (
            <g className={className} onMouseEnter={handleMouseEnter} onMouseLeave={handleMouseLeave} id={props.id}>
              <path d={path} className="shadow" />
              <path d={path} className="link" />
            </g>
          );
        }}
      </Spring>
    );
  },

  ensureSameLength: function(points) {
    // Spring needs constant list length, hoping that dagre will insert never more than 10
    const length = 10;
    let missing = length - points.length;

    while (missing) {
      points.unshift(points[0]);
      missing = length - points.length;
    }

    return points;
  },

  handleMouseEnter: function(ev) {
    AppActions.enterEdge(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveEdge(ev.currentTarget.id);
  }

});

module.exports = Edge;
