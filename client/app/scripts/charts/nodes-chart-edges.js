import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { hasSelectedNode as hasSelectedNodeFn } from '../utils/topology-utils';
import EdgeContainer from './edge-container';

class NodesChartEdges extends React.Component {
  render() {
    const { hasSelectedNode, highlightedEdgeIds, layoutEdges,
      layoutPrecision, searchNodeMatches = makeMap(), searchQuery,
      selectedNodeId } = this.props;

    return (
      <g className="nodes-chart-edges">
        {layoutEdges.toIndexedSeq().map(edge => {
          const sourceSelected = selectedNodeId === edge.get('source');
          const targetSelected = selectedNodeId === edge.get('target');
          const blurred = hasSelectedNode && !sourceSelected && !targetSelected
            || searchQuery && !(searchNodeMatches.has(edge.get('source'))
              && searchNodeMatches.has(edge.get('target')));
          const focused = hasSelectedNode && (sourceSelected || targetSelected);

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
              highlighted={highlightedEdgeIds.has(edge.get('id'))}
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
    selectedNodeId: state.get('selectedNodeId')
  };
}

export default connect(
  mapStateToProps
)(NodesChartEdges);
