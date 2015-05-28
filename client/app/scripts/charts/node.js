const React = require('react');
const tweenState = require('react-tween-state');

const NodeColorMixin = require('../mixins/node-color-mixin');

const Node = React.createClass({
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
    const transform = 'translate(' + this.getTweeningValue('x') + ',' + this.getTweeningValue('y') + ')';
    const scale = this.props.scale;
    const textOffsetX = 0;
    const textOffsetY = scale(0.5) + 18;
    const color = this.getNodeColor(this.props.label);
    const className = this.props.highlighted ? 'node highlighted' : 'node';

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
