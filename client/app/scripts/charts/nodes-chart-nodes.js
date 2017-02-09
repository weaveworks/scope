import React from 'react';
import { connect } from 'react-redux';
import { List as makeList } from 'immutable';

import { nodeMetricsSelector } from '../selectors/metrics';
import { currentTopologySearchNodeMatchesSelector } from '../selectors/search';
import { getAdjacentNodes } from '../utils/topology-utils';
import NodeContainer from './node-container';

class NodesChartNodes extends React.Component {
  render() {
    const { adjacentNodes, highlightedNodeIds, layoutNodes, isAnimated,
      mouseOverNodeId, nodeMetrics, selectedScale, searchQuery, selectedNetwork,
      selectedNodeId, searchNodeMatches } = this.props;

    // highlighter functions
    const setHighlighted = node => node.set('highlighted',
      highlightedNodeIds.has(node.get('id')) || selectedNodeId === node.get('id'));
    const setFocused = node => node.set('focused', selectedNodeId
      && (selectedNodeId === node.get('id')
      || (adjacentNodes && adjacentNodes.includes(node.get('id')))));
    const setBlurred = node => node.set('blurred',
      (selectedNodeId && !node.get('focused'))
      || (searchQuery && !searchNodeMatches.has(node.get('id')) && !node.get('highlighted'))
      || (selectedNetwork && !(node.get('networks') || makeList()).find(n => n.get('id') === selectedNetwork)));

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
        {nodesToRender.map(node => <NodeContainer
          blurred={node.get('blurred')}
          focused={node.get('focused')}
          matches={searchNodeMatches.get(node.get('id'))}
          highlighted={node.get('highlighted')}
          shape={node.get('shape')}
          networks={node.get('networks')}
          stack={node.get('stack')}
          key={node.get('id')}
          id={node.get('id')}
          label={node.get('label')}
          pseudo={node.get('pseudo')}
          subLabel={node.get('subLabel')}
          metric={nodeMetrics.get(node.get('id'))}
          rank={node.get('rank')}
          isAnimated={isAnimated}
          scale={node.get('focused') ? selectedScale : 1}
          dx={node.get('x')}
          dy={node.get('y')}
        />)}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    adjacentNodes: getAdjacentNodes(state),
    nodeMetrics: nodeMetricsSelector(state),
    highlightedNodeIds: state.get('highlightedNodeIds'),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    searchNodeMatches: currentTopologySearchNodeMatchesSelector(state),
    searchQuery: state.get('searchQuery'),
  };
}

export default connect(
  mapStateToProps,
)(NodesChartNodes);
