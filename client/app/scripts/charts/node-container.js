import React from 'react';
import { connect } from 'react-redux';
import { List as makeList } from 'immutable';
import { GraphNode } from 'weaveworks-ui-components';

import {
  getMetricValue,
  getMetricColor,
} from '../utils/metric-utils';
import { clickNode, enterNode, leaveNode } from '../actions/app-actions';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { getNodeColor } from '../utils/color-utils';
import MatchedResults from '../components/matched-results';
import { GRAPH_VIEW_MODE } from '../constants/naming';

import NodeNetworksOverlay from './node-networks-overlay';

class NodeContainer extends React.Component {
  saveRef = (ref) => {
    this.ref = ref;
  };

  handleMouseClick = (nodeId, ev) => {
    ev.stopPropagation();
    trackAnalyticsEvent('scope.node.click', {
      layout: GRAPH_VIEW_MODE,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
    this.props.clickNode(nodeId, this.props.label, this.ref.getBoundingClientRect());
  };

  renderPrependedInfo = () => {
    const { showingNetworks, networks } = this.props;
    if (!showingNetworks) return null;

    return (
      <NodeNetworksOverlay networks={networks} />
    );
  };

  renderAppendedInfo = () => {
    const matchedMetadata = this.props.matches.get('metadata', makeList());
    const matchedParents = this.props.matches.get('parents', makeList());
    const matchedDetails = matchedMetadata.concat(matchedParents);
    return (
      <MatchedResults matches={matchedDetails} />
    );
  };

  render() {
    const {
      rank, label, pseudo, metric, showingNetworks, networks
    } = this.props;
    const { hasMetric, height, formattedValue } = getMetricValue(metric);
    const metricFormattedValue = !pseudo && hasMetric ? formattedValue : '';
    const labelOffset = (showingNetworks && networks) ? 10 : 0;

    return (
      <GraphNode
        id={this.props.id}
        shape={this.props.shape}
        label={this.props.label}
        labelMinor={this.props.labelMinor}
        labelOffset={labelOffset}
        stacked={this.props.stacked}
        highlighted={this.props.highlighted}
        color={getNodeColor(rank, label, pseudo)}
        size={this.props.size}
        isAnimated={this.props.isAnimated}
        contrastMode={this.props.contrastMode}
        forceSvg={this.props.exportingGraph}
        searchTerms={this.props.searchTerms}
        metricColor={getMetricColor(metric)}
        metricFormattedValue={metricFormattedValue}
        metricNumericValue={height}
        renderPrependedInfo={this.renderPrependedInfo}
        renderAppendedInfo={this.renderAppendedInfo}
        onMouseEnter={this.props.enterNode}
        onMouseLeave={this.props.leaveNode}
        onClick={this.handleMouseClick}
        graphNodeRef={this.saveRef}
        x={this.props.x}
        y={this.props.y}
      />
    );
  }
}

function mapStateToProps(state) {
  return {
    searchTerms: [state.get('searchQuery')],
    exportingGraph: state.get('exportingGraph'),
    showingNetworks: state.get('showingNetworks'),
    currentTopology: state.get('currentTopology'),
    contrastMode: state.get('contrastMode'),
  };
}

export default connect(
  mapStateToProps,
  { clickNode, enterNode, leaveNode }
)(NodeContainer);
