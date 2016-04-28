import React from 'react';
import ReactDOM from 'react-dom';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { clickNode, enterNode, leaveNode } from '../actions/app-actions';
import { getNodeColor } from '../utils/color-utils';
import MatchedText from '../components/matched-text';
import MatchedResults from '../components/matched-results';

import NodeShapeCircle from './node-shape-circle';
import NodeShapeStack from './node-shape-stack';
import NodeShapeRoundedSquare from './node-shape-rounded-square';
import NodeShapeHex from './node-shape-hex';
import NodeShapeHeptagon from './node-shape-heptagon';
import NodeShapeCloud from './node-shape-cloud';

function stackedShape(Shape) {
  const factory = React.createFactory(NodeShapeStack);
  return props => factory(Object.assign({}, props, {shape: Shape}));
}

const nodeShapes = {
  circle: NodeShapeCircle,
  hexagon: NodeShapeHex,
  heptagon: NodeShapeHeptagon,
  square: NodeShapeRoundedSquare,
  cloud: NodeShapeCloud
};

function getNodeShape({ shape, stack }) {
  const nodeShape = nodeShapes[shape];
  if (!nodeShape) {
    throw new Error(`Unknown shape: ${shape}!`);
  }
  return stack ? stackedShape(nodeShape) : nodeShape;
}

function ellipsis(text, fontSize, maxWidth) {
  const averageCharLength = fontSize / 1.5;
  const allowedChars = maxWidth / averageCharLength;
  let truncatedText = text;
  if (text && text.length > allowedChars) {
    truncatedText = `${text.slice(0, allowedChars)}...`;
  }
  return truncatedText;
}

class Node extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
    this.state = { hovered: false };
  }

  render() {
    const { blurred, focused, highlighted, label, matched, matches, pseudo, rank,
      subLabel, scaleFactor, transform, zoomScale } = this.props;
    const { hovered } = this.state;
    const nodeScale = focused ? this.props.selectedNodeScale : this.props.nodeScale;

    const color = getNodeColor(rank, label, pseudo);
    const truncate = !focused && !hovered;
    const labelText = truncate ? ellipsis(label, 14, nodeScale(4 * scaleFactor)) : label;
    const subLabelText = truncate ? ellipsis(subLabel, 12, nodeScale(4 * scaleFactor)) : subLabel;

    let labelOffsetY = 18;
    let subLabelOffsetY = 35;
    let labelFontSize = 14;
    let subLabelFontSize = 12;

    // render focused nodes in normal size
    if (focused) {
      labelFontSize /= zoomScale;
      subLabelFontSize /= zoomScale;
      labelOffsetY /= zoomScale;
      subLabelOffsetY /= zoomScale;
    }

    const className = classNames({
      node: true,
      highlighted,
      blurred,
      hovered,
      matched,
      pseudo
    });

    const NodeShapeType = getNodeShape(this.props);

    return (
      <g className={className} transform={transform} onClick={this.handleMouseClick}
        onMouseEnter={this.handleMouseEnter} onMouseLeave={this.handleMouseLeave}>
        <rect className="hover-box"
          x={-nodeScale(scaleFactor * 0.5)}
          y={-nodeScale(scaleFactor * 0.5)}
          width={nodeScale(scaleFactor)}
          height={nodeScale(scaleFactor) + subLabelOffsetY}
          />
        <foreignObject x={-nodeScale(2 * scaleFactor)}
          y={labelOffsetY + nodeScale(0.5 * scaleFactor)}
          width={nodeScale(scaleFactor * 4)} height={subLabelOffsetY}>
          <div className="node-label" style={{fontSize: labelFontSize}}>
            <MatchedText text={labelText} matches={matches} fieldId="label" />
          </div>
          <div className="node-sublabel" style={{fontSize: subLabelFontSize}}>
            <MatchedText text={subLabelText} matches={matches} fieldId="sublabel" />
          </div>
          <MatchedResults matches={matches && matches.get('metadata')} />
        </foreignObject>
        <NodeShapeType
          size={nodeScale(scaleFactor)}
          color={color}
          {...this.props} />
      </g>
    );
  }

  handleMouseClick(ev) {
    ev.stopPropagation();
    this.props.clickNode(this.props.id, this.props.label,
      ReactDOM.findDOMNode(this).getBoundingClientRect());
  }

  handleMouseEnter() {
    this.props.enterNode(this.props.id);
    this.setState({ hovered: true });
  }

  handleMouseLeave() {
    this.props.leaveNode(this.props.id);
    this.setState({ hovered: false });
  }
}

export default connect(
  null,
  { clickNode, enterNode, leaveNode }
)(Node);
