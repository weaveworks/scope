var _ = require('lodash');
var React = require('react');
var tweenState = require('react-tween-state');

var NodeColorMixin = require('../mixins/node-color-mixin');

var Node = React.createClass({
  mixins: [
    NodeColorMixin,
    tweenState.Mixin
  ],

  getInitialState: function() {
    return {
      x: 0,
      y: 0
    };
  },

  componentWillMount: function() {
    // initial node position when rendered the first time
    this.setState({
      x: this.props.dx,
      y: this.props.dy
    });
  },

  componentWillReceiveProps: function(nextProps) {
    // animate node transition to next position
    this.tweenState('x', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.dx
    });
    this.tweenState('y', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.dy
    });
  },

  render: function() {
    var transform = "translate(" + this.getTweeningValue('x') + "," + this.getTweeningValue('y') + ")";
    var scale = this.props.scale;
    var textOffsetX = 0;
    var textOffsetY = scale(0.5) + 18;
    var textAngle = _.isUndefined(this.props.angle) ? 0 : -1 * (this.props.angle * 180 / Math.PI - 90);
    var color = this.getNodeColor(this.props.label);
    var className = this.props.highlighted ? "node highlighted" : "node";

    return (
      <g className={className} transform={transform} onClick={this.props.onClick} id={this.props.id}>
        <circle r={scale(0.5)} className="border" stroke={color}></circle>
        <circle r={scale(0.45)} className="shadow"></circle>
        <circle r={Math.max(2, scale(0.125))} className="node"></circle>
        <text className="node-label" textAnchor="middle" x={textOffsetX} y={textOffsetY}>{this.props.label}</text>
        <text className="node-sublabel" textAnchor="middle" x={textOffsetX} y={textOffsetY + 17}>{this.props.subLabel}</text>
      </g>
    );
  }
});

module.exports = Node;
