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
      (selectedNodeId && !node.get('focused'))
      || (searchQuery && !searchNodeMatches.has(node.get('id')) && !node.get('highlighted'))
      || (selectedNetwork && !(node.get('networks') || makeList()).find(n => n.get('id') === selectedNetwork)));

    // TODO: think about pulling this up into the store.
    const metric = (node) => {
      const isHighlighted = topCardNode && topCardNode.details && topCardNode.id === node.get('id');
      const sourceNode = isHighlighted ? fromJS(topCardNode.details) : node;
      return sourceNode.get('metrics') && sourceNode.get('metrics')
        .filter(m => m.get('id') === selectedMetric)
        .first();
    };

    const nodesToRender = layoutNodes.toIndexedSeq()
      .map(setHighlighted)
      .map(setFocused)
      .map(setBlurred);

    const renderNodeContainer = node => <NodeContainer
      highlighted={node.get('highlighted')}
      shape={node.get('shape')}
      key={node.get('id')}
      id={node.get('id')}
      rank={node.get('rank')}
      layoutPrecision={1}
      selectedNodeScale={nodeScale}
      nodeScale={nodeScale}
      dx={node.get('x')}
      dy={node.get('y')}
    />;

    const anyHighlighted = nodesToRender.some(node => node.get('highlighted'));
    return (
      <g className="nodes-chart-nodes">
        <g className="nodes-chart-nodes-background" style={{ opacity: anyHighlighted ? 0.2 : 1 }}>
          {nodesToRender.filter(node => !node.get('highlighted')).map(node => renderNodeContainer(node))}
        </g>
        <g className="nodes-chart-nodes-foreground">
          {nodesToRender.filter(node => node.get('highlighted')).map(node => renderNodeContainer(node))}
        </g>
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
