import React from 'react';
import { connect } from 'react-redux';
import { List as makeList, Map as makeMap } from 'immutable';

import { nodeMetricSelector } from '../selectors/node-metric';
import { searchNodeMatchesSelector } from '../selectors/search';
import { nodeNetworksSelector, selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { hasSelectedNode as hasSelectedNodeFn, getAdjacentNodes } from '../utils/topology-utils';
import NodeContainer from './node-container';
import EdgeContainer from './edge-container';

class NodesChartNodes extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.nodeRenderDecorator = this.nodeRenderDecorator.bind(this);
    this.nodeDisplayLayerDecorator = this.nodeDisplayLayerDecorator.bind(this);
    // Node decorators
    // TODO: Consider moving some of these one level up (or even to global selectors) so that
    // other components, like NodesChartEdges, could read more info directly from the nodes.
    this.nodeHighlightedDecorator = this.nodeHighlightedDecorator.bind(this);
    this.nodeFocusedDecorator = this.nodeFocusedDecorator.bind(this);
    this.nodeBlurredDecorator = this.nodeBlurredDecorator.bind(this);
    this.nodeMatchesDecorator = this.nodeMatchesDecorator.bind(this);
    this.nodeNetworksDecorator = this.nodeNetworksDecorator.bind(this);
    this.nodeMetricDecorator = this.nodeMetricDecorator.bind(this);
    this.nodeScaleDecorator = this.nodeScaleDecorator.bind(this);
    this.nodeHoveredDecorator = this.nodeHoveredDecorator.bind(this);
    this.nodeNormalDecorator = this.nodeNormalDecorator.bind(this);
    // Edge decorators
    this.edgeFocusedDecorator = this.edgeFocusedDecorator.bind(this);
    this.edgeBlurredDecorator = this.edgeBlurredDecorator.bind(this);
    this.edgeHighlightedDecorator = this.edgeHighlightedDecorator.bind(this);
    this.edgeScaleDecorator = this.edgeScaleDecorator.bind(this);
    this.edgeRenderDecorator = this.edgeRenderDecorator.bind(this);
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

  nodeHoveredDecorator(node) {
    return node.set('hovered', node.get('id') === this.props.mouseOverNodeId);
  }

  /* eslint class-methods-use-this: off */
  nodeNormalDecorator(node) {
    return node.set('normal',
      !(node.get('hovered') || node.get('blurred') || node.get('highlighted')));
  }

  // make sure blurred nodes are in the background
  nodeDisplayLayerDecorator(node) {
    let displayLayer;
    if (node.get('hovered')) {
      displayLayer = 'hovered';
    } else if (node.get('blurred') && !node.get('focused')) {
      displayLayer = 'blurred';
    } else if (node.get('highlighted')) {
      displayLayer = 'highlighted';
    } else {
      displayLayer = 'normal';
    }
    return node.set('displayLayer', displayLayer);
  }

  nodeRenderDecorator(node) {
    return node.set('render', () => (
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
        isAnimated={this.props.isAnimated}
        contrastMode={this.props.contrastMode}
      />
    ));
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

  edgeRenderDecorator(edge) {
    return edge.set('render', () => (
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
        isAnimated={this.props.isAnimated}
      />
    ));
  }

  render() {
    const nodesToRender = this.props.layoutNodes.toIndexedSeq()
      .map(this.nodeHighlightedDecorator)
      .map(this.nodeFocusedDecorator)
      .map(this.nodeBlurredDecorator)
      .map(this.nodeMatchesDecorator)
      .map(this.nodeNetworksDecorator)
      .map(this.nodeMetricDecorator)
      .map(this.nodeScaleDecorator)
      .map(this.nodeHoveredDecorator)
      .map(this.nodeNormalDecorator)
      .map(this.nodeDisplayLayerDecorator)
      .map(this.nodeRenderDecorator);

    const edgesToRender = this.props.layoutEdges.toIndexedSeq()
      .map(this.edgeHighlightedDecorator)
      .map(this.edgeFocusedDecorator)
      .map(this.edgeBlurredDecorator)
      .map(this.edgeScaleDecorator)
      .map(this.edgeRenderDecorator);

    const elementsToRender = makeList([
      edgesToRender.filter(edge => edge.get('blurred')),
      nodesToRender.filter(node => node.get('displayLayer') === 'blurred'),
      edgesToRender.filter(edge => !edge.get('blurred')),
      nodesToRender.filter(node => node.get('displayLayer') === 'normal'),
      nodesToRender.filter(node => node.get('displayLayer') === 'highlighted'),
      nodesToRender.filter(node => node.get('displayLayer') === 'hovered'),
    ]).flatten(true);
    // console.log(nodesToRenderWell.toJS());

    return (
      <g className="nodes-chart-nodes">
        {elementsToRender.map(n => n.get('render')())}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    hasSelectedNode: hasSelectedNodeFn(state),
    highlightedEdgeIds: state.get('highlightedEdgeIds'),
    nodeMetric: nodeMetricSelector(state),
    nodeNetworks: nodeNetworksSelector(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    neighborsOfSelectedNode: getAdjacentNodes(state),
    highlightedNodeIds: state.get('highlightedNodeIds'),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    searchQuery: state.get('searchQuery'),
    contrastMode: state.get('contrastMode')
  };
}

export default connect(
  mapStateToProps,
)(NodesChartNodes);
