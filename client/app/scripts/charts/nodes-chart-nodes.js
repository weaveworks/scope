import React from 'react';
import { connect } from 'react-redux';
import { fromJS, Map as makeMap, List as makeList } from 'immutable';

import { getAdjacentNodes } from '../utils/topology-utils';
import NodeContainer from './node-container';

class NodesChartNodes extends React.Component {
  render() {
    const { adjacentNodes, highlightedNodeIds, layoutNodes, layoutPrecision,
      mouseOverNodeId, nodeScale, scale, searchNodeMatches = makeMap(),
      searchQuery, selectedMetric, selectedNetwork, selectedNodeScale, selectedNodeId,
      topCardNode } = this.props;

    const zoomScale = scale;

    // highlighter functions
    const setHighlighted = node => node.set('highlighted',
      highlightedNodeIds.has(node.get('id')) || selectedNodeId === node.get('id'));
    const setFocused = node => node.set('focused', selectedNodeId
      && (selectedNodeId === node.get('id')
      || (adjacentNodes && adjacentNodes.includes(node.get('id')))));
    const setBlurred = node => node.set('blurred',
      selectedNodeId && !node.get('focused')
      || searchQuery && !searchNodeMatches.has(node.get('id'))
        && !node.get('highlighted')
      || selectedNetwork
        && !(node.get('networks') || makeList()).find(n => n.get('id') === selectedNetwork));

    // make sure blurred nodes are in the background
    const sortNodes = node => {
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

    // TODO: think about pulling this up into the store.
    const metric = node => {
      const isHighlighted = topCardNode && topCardNode.details && topCardNode.id === node.get('id');
      const sourceNode = isHighlighted ? fromJS(topCardNode.details) : node;
      return sourceNode.get('metrics') && sourceNode.get('metrics')
        .filter(m => m.get('id') === selectedMetric)
        .first();
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
          matched={searchNodeMatches.has(node.get('id'))}
          matches={searchNodeMatches.get(node.get('id'))}
          highlighted={node.get('highlighted')}
          shape={node.get('shape')}
          networks={node.get('networks')}
          stack={node.get('stack')}
          key={node.get('id')}
          id={node.get('id')}
          label={node.get('label')}
          pseudo={node.get('pseudo')}
          nodeCount={node.get('nodeCount')}
          subLabel={node.get('subLabel')}
          metric={metric(node)}
          rank={node.get('rank')}
          layoutPrecision={layoutPrecision}
          selectedNodeScale={selectedNodeScale}
          nodeScale={nodeScale}
          zoomScale={zoomScale}
          dx={node.get('x')}
          dy={node.get('y')} />)}
      </g>
    );
  }
}

function mapStateToProps(state) {
  const currentTopologyId = state.get('currentTopologyId');
  return {
    adjacentNodes: getAdjacentNodes(state),
    highlightedNodeIds: state.get('highlightedNodeIds'),
    mouseOverNodeId: state.get('mouseOverNodeId'),
    selectedMetric: state.get('selectedMetric'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
    searchNodeMatches: state.getIn(['searchNodeMatches', currentTopologyId]),
    searchQuery: state.get('searchQuery'),
    topCardNode: state.get('nodeDetails').last()
  };
}

export default connect(
  mapStateToProps
)(NodesChartNodes);
