const React = require('react');
const tweenState = require('react-tween-state');

const AppActions = require('../actions/app-actions');
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
    const isPseudo = !!this.props.pseudo;
    const color = isPseudo ? '' : this.getNodeColor(this.props.label);
    const onClick = this.props.onClick;
    const onMouseEnter = this.handleMouseEnter;
    const onMouseLeave = this.handleMouseLeave;
    const classNames = ['node'];

    if (this.props.highlighted) {
      classNames.push('highlighted');
    }

    if (this.props.pseudo) {
      classNames.push('pseudo');
    }

    return (
      <g className={classNames.join(' ')} transform={transform} id={this.props.id}
        onClick={onClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
        {this.props.highlighted && <circle r={scale(0.7)} className="highlighted"></circle>}
        <circle r={scale(0.5)} className="border" stroke={color}></circle>
        <circle r={scale(0.45)} className="shadow"></circle>
        <circle r={Math.max(2, scale(0.125))} className="node"></circle>
        <text className="node-label" textAnchor="middle" x={textOffsetX} y={textOffsetY}>
          {this.ellipsis(this.props.label, 14)}
        </text>
        <text className="node-sublabel" textAnchor="middle" x={textOffsetX} y={textOffsetY + 17}>
          {this.ellipsis(this.props.subLabel, 12)}
        </text>
      </g>
    );
  },

  ellipsis: function(text, fontSize) {
    const maxWidth = this.props.scale(4);
    const averageCharLength = fontSize / 1.5;
    const allowedChars = maxWidth / averageCharLength;
    let truncatedText = text;
    let trimmedText = text;
    while (text && trimmedText.length > 1 && trimmedText.length > allowedChars) {
      trimmedText = trimmedText.slice(0, -1);
      truncatedText = trimmedText + '...';
    }
    return truncatedText;
  },

  handleMouseEnter: function(ev) {
    AppActions.enterNode(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveNode(ev.currentTarget.id);
  }

});

module.exports = Node;
