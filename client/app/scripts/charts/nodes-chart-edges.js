import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { searchNodeMatchesSelector } from '../selectors/search';
import { selectedNetworkNodesIdsSelector } from '../selectors/node-networks';
import { hasSelectedNode as hasSelectedNodeFn } from '../utils/topology-utils';
import EdgeContainer from './edge-container';

class NodesChartEdges extends React.Component {
  constructor(props, context) {
    super(props, context);

    // Edge decorators
    this.edgeFocusedDecorator = this.edgeFocusedDecorator.bind(this);
    this.edgeBlurredDecorator = this.edgeBlurredDecorator.bind(this);
    this.edgeHighlightedDecorator = this.edgeHighlightedDecorator.bind(this);
    this.edgeScaleDecorator = this.edgeScaleDecorator.bind(this);
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

  render() {
    const { layoutEdges, isAnimated } = this.props;

    const edgesToRender = layoutEdges.toIndexedSeq()
      .map(this.edgeHighlightedDecorator)
      .map(this.edgeFocusedDecorator)
      .map(this.edgeBlurredDecorator)
      .map(this.edgeScaleDecorator);

    return (
      <g className="nodes-chart-edges">
        {edgesToRender.map(edge => (
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
        ))}
      </g>
    );
  }
}

export default connect(
  state => ({
    hasSelectedNode: hasSelectedNodeFn(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    selectedNetworkNodesIds: selectedNetworkNodesIdsSelector(state),
    searchQuery: state.get('searchQuery'),
    highlightedEdgeIds: state.get('highlightedEdgeIds'),
    selectedNetwork: state.get('selectedNetwork'),
    selectedNodeId: state.get('selectedNodeId'),
  })
)(NodesChartEdges);
