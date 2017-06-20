import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap, List as makeList } from 'immutable';

import NodeContainer from './node-container';
import EdgeContainer from './edge-container';
import { getAdjacentNodes, hasSelectedNode as hasSelectedNodeFn } from '../utils/topology-utils';
import { graphExceedsComplexityThreshSelector } from '../selectors/topology';
import { nodeNetworksSelector, selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { searchNodeMatchesSelector } from '../selectors/search';
import { nodeMetricSelector } from '../selectors/node-metric';
import {
  highlightedNodeIdsSelector,
  highlightedEdgeIdsSelector
} from '../selectors/graph-view/decorators';
import {
  selectedScaleSelector,
  layoutNodesSelector,
  layoutEdgesSelector
} from '../selectors/graph-view/layout';


const BLURRED_NODES_LAYER = 'blurred-nodes';
const NORMAL_NODES_LAYER = 'normal-nodes';
const HIGHLIGHTED_NODES_LAYER = 'highlighted-nodes';
const HOVERED_NODES_LAYER = 'hovered-nodes';

const NORMAL_EDGES_LAYER = 'normal-edges';

function edgeDisplayLayer() {
  return NORMAL_EDGES_LAYER;
}

class NodesChartElements extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.renderNode = this.renderNode.bind(this);
    this.renderEdge = this.renderEdge.bind(this);
    this.nodeDisplayLayer = this.nodeDisplayLayer.bind(this);
    // this.edgeDisplayLayer = this.edgeDisplayLayer.bind(this);

    // Node decorators
    this.nodeHighlightedDecorator = this.nodeHighlightedDecorator.bind(this);
    this.nodeFocusedDecorator = this.nodeFocusedDecorator.bind(this);
    this.nodeBlurredDecorator = this.nodeBlurredDecorator.bind(this);
    this.nodeMatchesDecorator = this.nodeMatchesDecorator.bind(this);
    this.nodeNetworksDecorator = this.nodeNetworksDecorator.bind(this);
    this.nodeMetricDecorator = this.nodeMetricDecorator.bind(this);
    this.nodeScaleDecorator = this.nodeScaleDecorator.bind(this);

    // Edge decorators
    this.edgeFocusedDecorator = this.edgeFocusedDecorator.bind(this);
    this.edgeBlurredDecorator = this.edgeBlurredDecorator.bind(this);
    this.edgeHighlightedDecorator = this.edgeHighlightedDecorator.bind(this);
    this.edgeScaleDecorator = this.edgeScaleDecorator.bind(this);
  }

  // make sure blurred nodes are in the background
  nodeDisplayLayer(node) {
    if (node.get('id') === this.props.mouseOverNodeId) {
      return HOVERED_NODES_LAYER;
    } else if (node.get('blurred') && !node.get('focused')) {
      return BLURRED_NODES_LAYER;
    } else if (node.get('highlighted')) {
      return HIGHLIGHTED_NODES_LAYER;
    }
    return NORMAL_NODES_LAYER;
  }

  nodeHighlightedDecorator(node) {
    const nodeSelected = (this.props.selectedNodeId === node.get('id'));
    const nodeHighlighted = this.props.highlightedNodeIds.has(node.get('id'));
    return node.set('highlighted', nodeHighlighted || nodeSelected);
  }

  nodeFocusedDecorator(node) {
    const nodeSelected = (this.props.selectedNodeId === node.get('id'));
    const isNeighborOfSelected = this.props.neighborsOfSelectedNode.includes(node.get('id'));
    return node.set('focused', nodeSelected || isNeighborOfSelected);
  }

  nodeBlurredDecorator(node) {
    const belongsToNetwork = this.props.selectedNetworkNodesIds.contains(node.get('id'));
    const noMatches = this.props.searchNodeMatches.get(node.get('id'), makeMap()).isEmpty();
    const notMatched = (this.props.searchQuery && !node.get('highlighted') && noMatches);
    const notFocused = (this.props.selectedNodeId && !node.get('focused'));
    const notInNetwork = (this.props.selectedNetwork && !belongsToNetwork);
    return node.set('blurred', notMatched || notFocused || notInNetwork);
  }

  nodeMatchesDecorator(node) {
    return node.set('matches', this.props.searchNodeMatches.get(node.get('id')));
  }

  nodeNetworksDecorator(node) {
    return node.set('networks', this.props.nodeNetworks.get(node.get('id')));
  }

  nodeMetricDecorator(node) {
    return node.set('metric', this.props.nodeMetric.get(node.get('id')));
  }

  nodeScaleDecorator(node) {
    return node.set('scale', node.get('focused') ? this.props.selectedScale : 1);
  }

  edgeHighlightedDecorator(edge) {
    return edge.set('highlighted', this.props.highlightedEdgeIds.has(edge.get('id')));
  }

  edgeFocusedDecorator(edge) {
    const sourceSelected = (this.props.selectedNodeId === edge.get('source'));
    const targetSelected = (this.props.selectedNodeId === edge.get('target'));
    return edge.set('focused', this.props.hasSelectedNode && (sourceSelected || targetSelected));
  }

  edgeBlurredDecorator(edge) {
    const { selectedNodeId, searchNodeMatches, selectedNetworkNodesIds } = this.props;
    const sourceSelected = (selectedNodeId === edge.get('source'));
    const targetSelected = (selectedNodeId === edge.get('target'));
    const otherNodesSelected = this.props.hasSelectedNode && !sourceSelected && !targetSelected;
    const sourceNoMatches = searchNodeMatches.get(edge.get('source'), makeMap()).isEmpty();
    const targetNoMatches = searchNodeMatches.get(edge.get('target'), makeMap()).isEmpty();
    const notMatched = this.props.searchQuery && (sourceNoMatches || targetNoMatches);
    const sourceInNetwork = selectedNetworkNodesIds.contains(edge.get('source'));
    const targetInNetwork = selectedNetworkNodesIds.contains(edge.get('target'));
    const notInNetwork = this.props.selectedNetwork && (!sourceInNetwork || !targetInNetwork);
    return edge.set('blurred', !edge.get('highlighted') && !edge.get('focused') &&
      (otherNodesSelected || notMatched || notInNetwork));
  }

  edgeScaleDecorator(edge) {
    return edge.set('scale', edge.get('focused') ? this.props.selectedScale : 1);
  }

  renderNode(node) {
    const { isAnimated, contrastMode } = this.props;
    return (
      <NodeContainer
        matches={node.get('matches')}
        networks={node.get('networks')}
        metric={node.get('metric')}
        blurred={node.get('blurred')}
        focused={node.get('focused')}
        highlighted={node.get('highlighted')}
        shape={node.get('shape')}
        stack={node.get('stack')}
        key={node.get('id')}
        id={node.get('id')}
        label={node.get('label')}
        labelMinor={node.get('labelMinor')}
        pseudo={node.get('pseudo')}
        rank={node.get('rank')}
        dx={node.get('x')}
        dy={node.get('y')}
        scale={node.get('scale')}
        isAnimated={isAnimated}
        contrastMode={contrastMode}
      />
    );
  }

  renderEdge(edge) {
    const { isAnimated } = this.props;
    return (
      <EdgeContainer
        key={edge.get('id')}
        id={edge.get('id')}
        source={edge.get('source')}
        target={edge.get('target')}
        waypoints={edge.get('points')}
        highlighted={edge.get('highlighted')}
        focused={edge.get('focused')}
        blurred={edge.get('blurred')}
        scale={edge.get('scale')}
        isAnimated={isAnimated}
      />
    );
  }

  render() {
    const nodes = this.props.layoutNodes.toIndexedSeq()
      .map(this.nodeHighlightedDecorator)
      .map(this.nodeFocusedDecorator)
      .map(this.nodeBlurredDecorator)
      .map(this.nodeMatchesDecorator)
      .map(this.nodeNetworksDecorator)
      .map(this.nodeMetricDecorator)
      .map(this.nodeScaleDecorator)
      .groupBy(this.nodeDisplayLayer);

    const edges = this.props.layoutEdges.toIndexedSeq()
      .map(this.edgeHighlightedDecorator)
      .map(this.edgeFocusedDecorator)
      .map(this.edgeBlurredDecorator)
      .map(this.edgeScaleDecorator)
      .groupBy(edgeDisplayLayer);

    console.log(nodes.toJS(), edges.toJS());
    return (
      <g className="nodes-chart-elements">
        {edges.get(NORMAL_EDGES_LAYER, makeList()).map(this.renderEdge)}
        {nodes.get(BLURRED_NODES_LAYER, makeList()).map(this.renderNode)}
        {nodes.get(NORMAL_NODES_LAYER, makeList()).map(this.renderNode)}
        {nodes.get(HIGHLIGHTED_NODES_LAYER, makeList()).map(this.renderNode)}
        {nodes.get(HOVERED_NODES_LAYER, makeList()).map(this.renderNode)}
      </g>
    );
  }
}


function mapStateToProps(state) {
  return {
    hasSelectedNode: hasSelectedNodeFn(state),
    layoutNodes: layoutNodesSelector(state),
    layoutEdges: layoutEdgesSelector(state),
    isAnimated: !graphExceedsComplexityThreshSelector(state),
    highlightedNodeIds: highlightedNodeIdsSelector(state),
    highlightedEdgeIds: highlightedEdgeIdsSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    neighborsOfSelectedNode: getAdjacentNodes(state),
    nodeNetworks: nodeNetworksSelector(state),
    nodeMetric: nodeMetricSelector(state),
    selectedScale: selectedScaleSelector(state),
    searchQuery: state.get('searchQuery'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    contrastMode: state.get('contrastMode'),
  };
}

export default connect(
  mapStateToProps
)(NodesChartElements);
