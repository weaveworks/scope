import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap, List as makeList } from 'immutable';

import { hasSelectedNode as hasSelectedNodeFn } from '../utils/topology-utils';
import EdgeContainer from './edge-container';

class NodesChartEdges extends React.Component {
  render() {
    const { hasSelectedNode, highlightedEdgeIds, layoutEdges,
      layoutPrecision, searchNodeMatches = makeMap(), searchQuery,
      selectedNodeId, selectedNetwork, selectedNetworkNodes } = this.props;

    return (
      <g className="nodes-chart-edges">
        {layoutEdges.toIndexedSeq().map((edge) => {
          const sourceSelected = selectedNodeId === edge.get('source');
          const targetSelected = selectedNodeId === edge.get('target');
          const highlighted = highlightedEdgeIds.has(edge.get('id'));
          if (!highlighted) return <span key={edge.get('id')} />;

          const focused = hasSelectedNode && (sourceSelected || targetSelected);
          const otherNodesSelected = hasSelectedNode && !sourceSelected && !targetSelected;
          const noMatches = searchQuery &&
            !(searchNodeMatches.has(edge.get('source')) &&
              searchNodeMatches.has(edge.get('target')));
          const noSelectedNetworks = selectedNetwork &&
            !(selectedNetworkNodes.contains(edge.get('source')) &&
              selectedNetworkNodes.contains(edge.get('target')));
          const blurred = !highlighted && (otherNodesSelected ||
                                           (!focused && noMatches) ||
                                             (!focused && noSelectedNetworks));

          return (
            <EdgeContainer
              key={edge.get('id')}
              id={edge.get('id')}
              source={edge.get('source')}
              target={edge.get('target')}
              points={edge.get('points')}
              blurred={blurred}
              focused={focused}
              layoutPrecision={layoutPrecision}
              highlighted={false}
            />
          );
        })}
      </g>
    );
  }
}

function mapStateToProps(state) {
  const currentTopologyId = state.get('currentTopologyId');
  return {
    hasSelectedNode: hasSelectedNodeFn(state),
    highlightedEdgeIds: state.get('highlightedEdgeIds'),
    searchNodeMatches: state.getIn(['searchNodeMatches', currentTopologyId]),
    searchQuery: state.get('searchQuery'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNetworkNodes: state.getIn(['networkNodes', state.get('selectedNetwork')], makeList()),
    selectedNodeId: state.get('selectedNodeId'),
  };
}

export default connect(
  mapStateToProps
)(NodesChartEdges);
