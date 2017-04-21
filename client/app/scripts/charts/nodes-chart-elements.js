import React from 'react';
import { connect } from 'react-redux';
import { List as makeList, Map as makeMap } from 'immutable';

import { nodeMetricSelector } from '../selectors/node-metric';
import { searchNodeMatchesSelector } from '../selectors/search';
import { nodeNetworksSelector, selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { hasSelectedNode as hasSelectedNodeFn, getAdjacentNodes } from '../utils/topology-utils';
import { graphExceedsComplexityThreshSelector } from '../selectors/topology';
import {
  selectedScaleSelector,
  layoutNodesSelector,
  layoutEdgesSelector
} from '../selectors/graph-view/layout';
import {
  highlightedNodeIdsSelector,
  focusedNodeIdsSelector,
  blurredNodeIdsSelector,
} from '../selectors/graph-view/nodes';
import {
  highlightedEdgeIdsSelector,
  focusedEdgeIdsSelector,
} from '../selectors/graph-view/edges';
import NodeContainer from './node-container';
import EdgeContainer from './edge-container';


class NodesChartElements extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.nodeRenderDecorator = this.nodeRenderDecorator.bind(this);
    this.nodeDisplayLayerDecorator = this.nodeDisplayLayerDecorator.bind(this);
    // Node decorators
    // TODO: Consider moving some of these one level up (or even to global selectors) so that
    // other components, like NodesChartEdges, could read more info directly from the nodes.
    this.nodeScaleDecorator = this.nodeScaleDecorator.bind(this);
    this.nodeHoveredDecorator = this.nodeHoveredDecorator.bind(this);
    // Edge decorators
    this.edgeBlurredDecorator = this.edgeBlurredDecorator.bind(this);
    this.edgeScaleDecorator = this.edgeScaleDecorator.bind(this);
    this.edgeDisplayLayerDecorator = this.edgeDisplayLayerDecorator.bind(this);
    this.edgeRenderDecorator = this.edgeRenderDecorator.bind(this);
  }

  nodeScaleDecorator(node) {
    return node.set('scale', this.props.focusedNodeIds.has(node.get('id')) ? this.props.selectedScale : 1);
  }

  nodeHoveredDecorator(node) {
    return node.set('hovered', node.get('id') === this.props.mouseOverNodeId);
  }

  /* eslint class-methods-use-this: off */
  // make sure blurred nodes are in the background
  nodeDisplayLayerDecorator(node) {
    const id = node.get('id');
    let displayLayer;

    if (this.props.mouseOverNodeId === id) {
      displayLayer = 'hovered';
    } else if (this.props.blurredNodeIds.has(id) && !this.props.highlightedNodeIds.has(id)) {
      displayLayer = 'blurred';
    } else if (this.props.highlightedNodeIds.has(id)) {
      displayLayer = 'highlighted';
    } else {
      displayLayer = 'normal';
    }

    return node.set('displayLayer', displayLayer);
  }

  nodeRenderDecorator(node) {
    const id = node.get('id');
    return node.set('render', () => (
      <NodeContainer
        id={id}
        key={id}
        matches={this.props.searchNodeMatches.get(id)}
        networks={this.props.nodeNetworks.get(id)}
        metric={this.props.nodeMetric.get(id)}
        blurred={this.props.blurredNodeIds.has(id)}
        focused={this.props.focusedNodeIds.has(id)}
        highlighted={this.props.highlightedNodeIds.has(id)}
        shape={node.get('shape')}
        stack={node.get('stack')}
        label={node.get('label')}
        labelMinor={node.get('labelMinor')}
        pseudo={node.get('pseudo')}
        rank={node.get('rank')}
        dx={node.get('x')}
        dy={node.get('y')}
        scale={node.get('scale')}
        isAnimated={this.props.isAnimated}
        contrastMode={this.props.contrastMode}
      />
    ));
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

  edgeDisplayLayerDecorator(edge) {
    let displayLayer;
    if (edge.get('blurred')) {
      displayLayer = 'blurred';
    } else {
      displayLayer = 'normal';
    }
    return edge.set('displayLayer', displayLayer);
  }

  edgeRenderDecorator(edge) {
    const id = edge.get('id');
    return edge.set('render', () => (
      <EdgeContainer
        id={id}
        key={id}
        source={edge.get('source')}
        target={edge.get('target')}
        waypoints={edge.get('points')}
        highlighted={this.props.highlightedEdgeIds.has(id)}
        focused={this.props.focusedEdgeIds.has(id)}
        blurred={edge.get('blurred')}
        scale={edge.get('scale')}
        isAnimated={this.props.isAnimated}
      />
    ));
  }

  render() {
    const nodesToRender = this.props.layoutNodes.toIndexedSeq()
      .map(this.nodeScaleDecorator)
      .map(this.nodeHoveredDecorator)
      .map(this.nodeDisplayLayerDecorator)
      .map(this.nodeRenderDecorator);

    const edgesToRender = this.props.layoutEdges.toIndexedSeq()
      .map(this.edgeBlurredDecorator)
      .map(this.edgeScaleDecorator)
      .map(this.edgeDisplayLayerDecorator)
      .map(this.edgeRenderDecorator);

    const elementsToRender = makeList([
      edgesToRender.filter(edge => edge.get('displayLayer') === 'blurred'),
      nodesToRender.filter(node => node.get('displayLayer') === 'blurred'),
      edgesToRender.filter(edge => edge.get('displayLayer') === 'normal'),
      nodesToRender.filter(node => node.get('displayLayer') === 'normal'),
      nodesToRender.filter(node => node.get('displayLayer') === 'highlighted'),
      nodesToRender.filter(node => node.get('displayLayer') === 'hovered'),
    ]).flatten(true);

    return (
      <g className="nodes-chart-elements">
        {elementsToRender.map(n => n.get('render')())}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    layoutNodes: layoutNodesSelector(state),
    layoutEdges: layoutEdgesSelector(state),
    selectedScale: selectedScaleSelector(state),
    isAnimated: !graphExceedsComplexityThreshSelector(state),
    hasSelectedNode: hasSelectedNodeFn(state),
    highlightedEdgeIds: highlightedEdgeIdsSelector(state),
    focusedEdgeIds: focusedEdgeIdsSelector(state),
    nodeMetric: nodeMetricSelector(state),
    nodeNetworks: nodeNetworksSelector(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    neighborsOfSelectedNode: getAdjacentNodes(state),
    highlightedNodeIds: highlightedNodeIdsSelector(state),
    focusedNodeIds: focusedNodeIdsSelector(state),
    blurredNodeIds: blurredNodeIdsSelector(state),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    searchQuery: state.get('searchQuery'),
    contrastMode: state.get('contrastMode'),
  };
}

export default connect(
  mapStateToProps,
)(NodesChartElements);
