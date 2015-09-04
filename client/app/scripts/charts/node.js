const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');
const NodeColorMixin = require('../mixins/node-color-mixin');

const Node = React.createClass({
  mixins: [
    NodeColorMixin
  ],

  getInitialState: function() {
    return {
      x: 0,
      y: 0
    };
  },

  render: function() {
    const props = this.props;
    const scale = this.props.scale;
    const textOffsetX = 0;
    const textOffsetY = scale(0.5) + 18;
    const isPseudo = !!this.props.pseudo;
    const color = isPseudo ? '' : this.getNodeColor(this.props.rank);
    const onClick = this.props.onClick;
    const onMouseEnter = this.handleMouseEnter;
    const onMouseLeave = this.handleMouseLeave;
    const classNames = ['node'];
    const ellipsis = this.ellipsis;

    if (this.props.highlighted) {
      classNames.push('highlighted');
    }

    if (this.props.pseudo) {
      classNames.push('pseudo');
    }

    return (
      <Spring endValue={{x: this.props.dx, y: this.props.dy}}>
        {function(interpolated) {
          const transform = 'translate(' + interpolated.x + ',' + interpolated.y + ')';
          return (
            <g className={classNames.join(' ')} transform={transform} id={props.id}
              onClick={onClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
              {props.highlighted && <circle r={scale(0.7)} className="highlighted"></circle>}
              <circle r={scale(0.5)} className="border" stroke={color}></circle>
              <circle r={scale(0.45)} className="shadow"></circle>
              <circle r={Math.max(2, scale(0.125))} className="node"></circle>
              <text className="node-label" textAnchor="middle" x={textOffsetX} y={textOffsetY}>
                {ellipsis(props.label, 14)}
              </text>
              <text className="node-sublabel" textAnchor="middle" x={textOffsetX} y={textOffsetY + 17}>
                {ellipsis(props.subLabel, 12)}
              </text>
            </g>
          );
        }}
      </Spring>
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
