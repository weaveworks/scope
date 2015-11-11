const _ = require('lodash');
const d3 = require('d3');
const React = require('react');
const Motion = require('react-motion').Motion;
const spring = require('react-motion').spring;

const AppActions = require('../actions/app-actions');

const line = d3.svg.line()
  .interpolate('basis')
  .x(function(d) { return d.x; })
  .y(function(d) { return d.y; });

const animConfig = [80, 20]; // stiffness, bounce

const flattenPoints = function(points) {
  const flattened = {};
  points.forEach(function(point, i) {
    flattened['x' + i] = spring(point.x, animConfig);
    flattened['y' + i] = spring(point.y, animConfig);
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
    extracted[index][axis] = value;
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
    const classNames = ['edge'];
    const points = flattenPoints(this.props.points);
    const props = this.props;
    const handleMouseEnter = this.handleMouseEnter;
    const handleMouseLeave = this.handleMouseLeave;

    if (this.props.highlighted) {
      classNames.push('highlighted');
    }
    if (this.props.blurred) {
      classNames.push('blurred');
    }
    const classes = classNames.join(' ');

    return (
      <Motion style={points}>
        {function(interpolated) {
          const path = line(extractPoints(interpolated));
          return (
            <g className={classes} onMouseEnter={handleMouseEnter} onMouseLeave={handleMouseLeave} id={props.id}>
              <path d={path} className="shadow" />
              <path d={path} className="link" />
            </g>
          );
        }}
      </Motion>
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
