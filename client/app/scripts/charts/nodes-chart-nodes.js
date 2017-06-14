import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { nodeMetricSelector } from '../selectors/node-metric';
import { searchNodeMatchesSelector } from '../selectors/search';
import { highlightedNodeIdsSelector } from '../selectors/graph-view/decorators';
import { nodeNetworksSelector, selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { getAdjacentNodes } from '../utils/topology-utils';
import NodeContainer from './node-container';

class NodesChartNodes extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.nodeDisplayLayer = this.nodeDisplayLayer.bind(this);
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

  // make sure blurred nodes are in the background
  nodeDisplayLayer(node) {
    if (node.get('id') === this.props.mouseOverNodeId) {
      return 3;
    }
    if (node.get('blurred') && !node.get('focused')) {
      return 0;
    }
    if (node.get('highlighted')) {
      return 2;
    }
    return 1;
  }

  render() {
    const { layoutNodes, isAnimated, contrastMode } = this.props;

    const nodesToRender = layoutNodes.toIndexedSeq()
      .map(this.nodeHighlightedDecorator)
      .map(this.nodeFocusedDecorator)
      .map(this.nodeBlurredDecorator)
      .map(this.nodeMatchesDecorator)
      .map(this.nodeNetworksDecorator)
      .map(this.nodeMetricDecorator)
      .map(this.nodeScaleDecorator)
      .sortBy(this.nodeDisplayLayer);

    return (
      <g className="nodes-chart-nodes">
        {nodesToRender.map(node => (
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
        ))}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    nodeMetric: nodeMetricSelector(state),
    nodeNetworks: nodeNetworksSelector(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    neighborsOfSelectedNode: getAdjacentNodes(state),
    highlightedNodeIds: highlightedNodeIdsSelector(state),
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
