import React from 'react';
import { connect } from 'react-redux';
import classnames from 'classnames';
import { Map as makeMap, List as makeList } from 'immutable';

import { clickNode, enterNode, leaveNode } from '../actions/app-actions';
import { getNodeColor } from '../utils/color-utils';
import MatchedText from '../components/matched-text';
import MatchedResults from '../components/matched-results';

import NodeShapeCircle from './node-shape-circle';
import NodeShapeStack from './node-shape-stack';
import NodeShapeRoundedSquare from './node-shape-rounded-square';
import NodeShapeHexagon from './node-shape-hexagon';
import NodeShapeHeptagon from './node-shape-heptagon';
import NodeShapeCloud from './node-shape-cloud';
import NodeNetworksOverlay from './node-networks-overlay';
import { MIN_NODE_LABEL_SIZE, BASE_NODE_LABEL_SIZE, BASE_NODE_SIZE } from '../constants/styles';


function labelFontSize(nodeSize) {
  return Math.max(MIN_NODE_LABEL_SIZE, (BASE_NODE_LABEL_SIZE / BASE_NODE_SIZE) * nodeSize);
}

function stackedShape(Shape) {
  const factory = React.createFactory(NodeShapeStack);
  return props => factory(Object.assign({}, props, {shape: Shape}));
}

const nodeShapes = {
  circle: NodeShapeCircle,
  hexagon: NodeShapeHexagon,
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

function svgLabels(label, subLabel, labelClassName, subLabelClassName, labelOffsetY) {
  return (
    <g className="node-label-svg">
      <text className={labelClassName} y={labelOffsetY + 18} textAnchor="middle">{label}</text>
      <text className={subLabelClassName} y={labelOffsetY + 35} textAnchor="middle">
        {subLabel}
      </text>
    </g>
  );
}

class Node extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
    this.saveShapeRef = this.saveShapeRef.bind(this);
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
    const { blurred, focused, highlighted, label, matches = makeMap(), networks,
      pseudo, rank, subLabel, scaleFactor, transform, exportingGraph,
      showingNetworks, stack } = this.props;
    const { hovered, matched } = this.state;
    const nodeScale = focused ? this.props.selectedNodeScale : this.props.nodeScale;

    const color = getNodeColor(rank, label, pseudo);
    const truncate = !focused && !hovered;
    const labelWidth = nodeScale(scaleFactor * 3);
    const labelOffsetX = -labelWidth / 2;
    const labelDy = (showingNetworks && networks) ? 0.70 : 0.55;
    const labelOffsetY = nodeScale(labelDy * scaleFactor);
    const networkOffset = nodeScale(scaleFactor * 0.67);

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
    const useSvgLabels = exportingGraph;
    const size = nodeScale(scaleFactor);
    const fontSize = labelFontSize(size);
    const mouseEvents = {
      onClick: this.handleMouseClick,
      onMouseEnter: this.handleMouseEnter,
      onMouseLeave: this.handleMouseLeave,
    };
    const matchedNodeDetails = matches.get('metadata', makeList())
      .concat(matches.get('parents', makeList()));

    return (
      <g className={nodeClassName} transform={transform}>

        {useSvgLabels ?

          svgLabels(label, subLabel, labelClassName, subLabelClassName, labelOffsetY) :

          <foreignObject
            style={{pointerEvents: 'none'}}
            x={labelOffsetX} y={labelOffsetY}
            width={labelWidth} height="100em">
            <div
              className="node-label-wrapper"
              style={{pointerEvents: 'all', fontSize, maxWidth: labelWidth}}
              {...mouseEvents}>
              <div className={labelClassName}>
                <MatchedText text={label} match={matches.get('label')} />
              </div>
              <div className={subLabelClassName}>
                <MatchedText text={subLabel} match={matches.get('sublabel')} />
              </div>
              {!blurred && <MatchedResults matches={matchedNodeDetails} />}
            </div>
          </foreignObject>}

        <g {...mouseEvents} ref={this.saveShapeRef}>
          <NodeShapeType
            size={size}
            color={color}
            {...this.props} />
        </g>

        {showingNetworks && <NodeNetworksOverlay
          offset={networkOffset}
          size={size} networks={networks}
          stack={stack}
        />}
      </g>
    );
  }

  saveShapeRef(ref) {
    this.shapeRef = ref;
  }

  handleMouseClick(ev) {
    ev.stopPropagation();
    this.props.clickNode(this.props.id, this.props.label, this.shapeRef.getBoundingClientRect());
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
  state => ({
    searchQuery: state.get('searchQuery'),
    exportingGraph: state.get('exportingGraph'),
    showingNetworks: state.get('showingNetworks'),
  }),
  { clickNode, enterNode, leaveNode }
)(Node);
