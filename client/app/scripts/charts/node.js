import React from 'react';
import ReactDOM from 'react-dom';
import { connect } from 'react-redux';
import classnames from 'classnames';
import { Map as makeMap } from 'immutable';

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

class Node extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
    this.state = {
      hovered: false,
      matched: false
    };
  }

  componentWillReceiveProps(nextProps) {
    // marks as matched only when search query changes
    if (nextProps.searchQuery !== this.props.searchQuery) {
      this.setState({
        matched: nextProps.matched
      });
    } else {
      this.setState({
        matched: false
      });
    }
  }

  render() {
    const { blurred, focused, highlighted, label, matches = makeMap(),
      pseudo, rank, subLabel, scaleFactor, transform, zoomScale } = this.props;
    const { hovered, matched } = this.state;
    const nodeScale = focused ? this.props.selectedNodeScale : this.props.nodeScale;

    const color = getNodeColor(rank, label, pseudo);
    const truncate = !focused && !hovered;
    const labelTransform = focused ? `scale(${1 / zoomScale})` : '';
    const labelWidth = nodeScale(scaleFactor * 4);
    const labelOffsetX = -labelWidth / 2;
    const labelOffsetY = focused ? nodeScale(0.5) : nodeScale(0.5 * scaleFactor);

    const nodeClassName = classnames('node', {
      highlighted,
      blurred: blurred && !focused,
      hovered,
      matched,
      pseudo
    });

    const labelClassName = classnames('node-label', { truncate });
    const subLabelClassName = classnames('node-sublabel', { truncate });

    const NodeShapeType = getNodeShape(this.props);

    return (
      <g className={nodeClassName} transform={transform}
        onMouseEnter={this.handleMouseEnter} onMouseLeave={this.handleMouseLeave}>
        {/* For browser */}
        <foreignObject x={labelOffsetX} y={labelOffsetY} width={labelWidth} height="10em"
          transform={labelTransform}>
          <div className="node-label-wrapper" onClick={this.handleMouseClick}>
            <div className={labelClassName}>
              <MatchedText text={label} match={matches.get('label')} />
            </div>
            <div className={subLabelClassName}>
              <MatchedText text={subLabel} match={matches.get('sublabel')} />
            </div>
            {!blurred && <MatchedResults matches={matches.get('metadata')} />}
          </div>
        </foreignObject>
        {/* For SVG export */}
        <g className="node-label-svg">
          <text className={labelClassName} y={labelOffsetY + 18} textAnchor="middle">{label}</text>
          <text className={subLabelClassName} y={labelOffsetY + 35} textAnchor="middle">
            {subLabel}
          </text>
        </g>
        <g onClick={this.handleMouseClick}>
          <NodeShapeType
            size={nodeScale(scaleFactor)}
            color={color}
            {...this.props} />
        </g>
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
  state => ({ searchQuery: state.get('searchQuery') }),
  { clickNode, enterNode, leaveNode }
)(Node);
