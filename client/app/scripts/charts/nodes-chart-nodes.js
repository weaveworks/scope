import React from 'react';
import { connect } from 'react-redux';

import { nodeNetworksSelector, selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { nodeMetricSelector } from '../selectors/node-metric';
import { currentTopologySearchNodeMatchesSelector } from '../selectors/search';
import { getAdjacentNodes } from '../utils/topology-utils';
import NodeContainer from './node-container';

class NodesChartNodes extends React.Component {
  render() {
    const { adjacentNodes, highlightedNodeIds, layoutNodes, isAnimated,
      mouseOverNodeId, nodeMetric, selectedScale, searchQuery, selectedNetwork,
      selectedNodeId, searchNodeMatches, nodeNetworks, selectedNetworkNodesIds } = this.props;

    // highlighter functions
    const setHighlighted = node => node.set('highlighted',
      highlightedNodeIds.has(node.get('id')) || selectedNodeId === node.get('id'));
    const setFocused = node => node.set('focused', selectedNodeId
      && (selectedNodeId === node.get('id')
      || (adjacentNodes && adjacentNodes.includes(node.get('id')))));
    const setBlurred = node => node.set('blurred',
      (selectedNodeId && !node.get('focused'))
      || (searchQuery && !searchNodeMatches.has(node.get('id')) && !node.get('highlighted'))
      || (selectedNetwork && !selectedNetworkNodesIds.contains(node.get('id'))));

    // make sure blurred nodes are in the background
    const sortNodes = (node) => {
      if (node.get('id') === mouseOverNodeId) {
        return 3;
      }
      if (node.get('blurred') && !node.get('focused')) {
        return 0;
      }
      if (node.get('highlighted')) {
        return 2;
      }
      return 1;
    };

    const nodesToRender = layoutNodes.toIndexedSeq()
      .map(setHighlighted)
      .map(setFocused)
      .map(setBlurred)
      .sortBy(sortNodes);

    return (
      <g className="nodes-chart-nodes">
        {nodesToRender.map((node) => {
          const nodeScale = node.get('focused') ? selectedScale : 1;
          const nodeId = node.get('id');
          return (
            <NodeContainer
              matches={searchNodeMatches.get(nodeId)}
              networks={nodeNetworks.get(nodeId)}
              metric={nodeMetric.get(nodeId)}
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
              scale={nodeScale}
              isAnimated={isAnimated}
            />
          );
        })}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    adjacentNodes: getAdjacentNodes(state),
    nodeMetric: nodeMetricSelector(state),
    nodeNetworks: nodeNetworksSelector(state),
    searchNodeMatches: currentTopologySearchNodeMatchesSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    highlightedNodeIds: state.get('highlightedNodeIds'),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    searchQuery: state.get('searchQuery'),
  };
}

export default connect(
  mapStateToProps,
)(NodesChartNodes);
