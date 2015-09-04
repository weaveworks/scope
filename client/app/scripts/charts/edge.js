const _ = require('lodash');
const d3 = require('d3');
const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');

const line = d3.svg.line()
  .interpolate('basis')
  .x(function(d) { return d.x; })
  .y(function(d) { return d.y; });

const flattenPoints = function(points) {
  const flattened = {};
  points.forEach(function(point, i) {
    flattened['x' + i] = point.x;
    flattened['y' + i] = point.y;
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
    return flattenPoints(this.props.points);
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

  handleMouseEnter: function(ev) {
    AppActions.enterEdge(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveEdge(ev.currentTarget.id);
  }

});

module.exports = Edge;
