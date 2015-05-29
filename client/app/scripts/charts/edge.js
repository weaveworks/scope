const d3 = require('d3');
const React = require('react');

const AppActions = require('../actions/app-actions');

const line = d3.svg.line()
  .interpolate('basis')
  .x(function(d) { return d.x; })
  .y(function(d) { return d.y; });

const Edge = React.createClass({

  render: function() {
    const className = this.props.highlighted ? 'edge highlighted' : 'edge';

    return (
      <g className={className} onMouseEnter={this.handleMouseEnter} onMouseLeave={this.handleMouseLeave} id={this.props.id}>
        <path d={line(this.props.points)} className="shadow" />
        <path d={line(this.props.points)} className="link" />
      </g>
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
