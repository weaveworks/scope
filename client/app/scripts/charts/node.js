const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');
const NodeColorMixin = require('../mixins/node-color-mixin');

const Node = React.createClass({
  mixins: [
    NodeColorMixin
  ],

  render: function() {
    const props = this.props;
    const scale = this.props.scale;
    let scaleFactor = 1;
    if (props.focused) {
      scaleFactor = 1.25;
    } else if (props.blurred) {
      scaleFactor = 0.75;
    }
    const labelOffsetY = 18;
    const subLabelOffsetY = labelOffsetY + 17;
    const isPseudo = !!this.props.pseudo;
    const color = isPseudo ? '' : this.getNodeColor(this.props.rank);
    const onClick = this.props.onClick;
    const onMouseEnter = this.handleMouseEnter;
    const onMouseLeave = this.handleMouseLeave;
    const classNames = ['node'];
    const animConfig = [80, 20]; // stiffness, bounce
    const label = this.ellipsis(props.label, 14, scale(4 * scaleFactor));
    const subLabel = this.ellipsis(props.subLabel, 12, scale(4 * scaleFactor));

    if (props.highlighted) {
      classNames.push('highlighted');
    }
    if (this.props.blurred) {
      classNames.push('blurred');
    }
    if (this.props.pseudo) {
      classNames.push('pseudo');
    }
    const classes = classNames.join(' ');

    return (
      <Spring endValue={{
        x: {val: this.props.dx, config: animConfig},
        y: {val: this.props.dy, config: animConfig},
        f: {val: scaleFactor, config: animConfig}
      }}>
        {function(interpolated) {
          const transform = `translate(${interpolated.x.val},${interpolated.y.val})`;
          return (
            <g className={classes} transform={transform} id={props.id}
              onClick={onClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
              {props.highlighted && <circle r={scale(0.7 * interpolated.f.val)} className="highlighted"></circle>}
              <circle r={scale(0.5 * interpolated.f.val)} className="border" stroke={color}></circle>
              <circle r={scale(0.45 * interpolated.f.val)} className="shadow"></circle>
              <circle r={Math.max(2, scale(0.125 * interpolated.f.val))} className="node"></circle>
              <text className="node-label" textAnchor="middle" x="0" y={labelOffsetY + scale(0.5 * interpolated.f.val)}>
                {label}
              </text>
              <text className="node-sublabel" textAnchor="middle" x="0" y={subLabelOffsetY + scale(0.5 * interpolated.f.val)}>
                {subLabel}
              </text>
            </g>
          );
        }}
      </Spring>
    );
  },

  ellipsis: function(text, fontSize, maxWidth) {
    const averageCharLength = fontSize / 1.5;
    const allowedChars = maxWidth / averageCharLength;
    let truncatedText = text;
    if (text && text.length > allowedChars) {
      truncatedText = text.slice(0, allowedChars) + '...';
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
