const React = require('react');
const Motion = require('react-motion').Motion;
const spring = require('react-motion').spring;

const AppActions = require('../actions/app-actions');
const NodeColorMixin = require('../mixins/node-color-mixin');

const Node = React.createClass({
  mixins: [
    NodeColorMixin
  ],

  render: function() {
    const props = this.props;
    const nodeScale = this.props.nodeScale;
    const zoomScale = this.props.zoomScale;
    let scaleFactor = 1;
    if (props.focused) {
      scaleFactor = 1.25 / zoomScale;
    } else if (props.blurred) {
      scaleFactor = 0.75;
    }
    let labelOffsetY = 18;
    let subLabelOffsetY = 35;
    const isPseudo = !!this.props.pseudo;
    const color = isPseudo ? '' : this.getNodeColor(this.props.rank);
    const onMouseEnter = this.handleMouseEnter;
    const onMouseLeave = this.handleMouseLeave;
    const onMouseClick = this.handleMouseClick;
    const classNames = ['node'];
    const animConfig = [80, 20]; // stiffness, bounce
    const label = this.ellipsis(props.label, 14, nodeScale(4 * scaleFactor));
    const subLabel = this.ellipsis(props.subLabel, 12, nodeScale(4 * scaleFactor));
    let labelFontSize = 14;
    let subLabelFontSize = 12;

    if (props.focused) {
      labelFontSize /= zoomScale;
      subLabelFontSize /= zoomScale;
      labelOffsetY /= zoomScale;
      subLabelOffsetY /= zoomScale;
    }
    if (this.props.highlighted) {
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
      <Motion style={{
        x: spring(this.props.dx, animConfig),
        y: spring(this.props.dy, animConfig),
        f: spring(scaleFactor, animConfig),
        labelFontSize: spring(labelFontSize, animConfig),
        subLabelFontSize: spring(subLabelFontSize, animConfig),
        labelOffsetY: spring(labelOffsetY, animConfig),
        subLabelOffsetY: spring(subLabelOffsetY, animConfig)
      }}>
        {function(interpolated) {
          const transform = `translate(${interpolated.x},${interpolated.y})`;
          return (
            <g className={classes} transform={transform} id={props.id}
              onClick={onMouseClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
              {props.highlighted && <circle r={nodeScale(0.7 * interpolated.f)} className="highlighted"></circle>}
              <circle r={nodeScale(0.5 * interpolated.f)} className="border" stroke={color}></circle>
              <circle r={nodeScale(0.45 * interpolated.f)} className="shadow"></circle>
              <circle r={Math.max(2, nodeScale(0.125 * interpolated.f))} className="node"></circle>
              <text className="node-label" textAnchor="middle" style={{fontSize: interpolated.labelFontSize}}
                x="0" y={interpolated.labelOffsetY + nodeScale(0.5 * interpolated.f)}>
                {label}
              </text>
              <text className="node-sublabel" textAnchor="middle" style={{fontSize: interpolated.subLabelFontSize}}
                x="0" y={interpolated.subLabelOffsetY + nodeScale(0.5 * interpolated.f)}>
                {subLabel}
              </text>
            </g>
          );
        }}
      </Motion>
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

  handleMouseClick: function(ev) {
    ev.stopPropagation();
    AppActions.clickNode(ev.currentTarget.id);
  },

  handleMouseEnter: function(ev) {
    AppActions.enterNode(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveNode(ev.currentTarget.id);
  }

});

module.exports = Node;
