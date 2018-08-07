import React from 'react';
import { connect } from 'react-redux';
import { GraphNode } from 'weaveworks-ui-components';

import { clickNode, enterNode, leaveNode } from '../actions/app-actions';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { getNodeColor } from '../utils/color-utils';
import { GRAPH_VIEW_MODE } from '../constants/naming';

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

  render() {
    const { rank, label, pseudo } = this.props;
    console.log('rerender');

    return (
      <GraphNode
        id={this.props.id}
        shape={this.props.shape}
        label={this.props.label}
        labelMinor={this.props.labelMinor}
        stacked={this.props.stacked}
        highlighted={this.props.highlighted}
        color={getNodeColor(rank, label, pseudo)}
        size={this.props.size}
        isAnimated={this.props.isAnimated}
        contrastMode={this.props.contrastMode}
        forceSvg={this.props.exportingGraph}
        graphNodeRef={this.saveRef}
        onMouseEnter={this.props.enterNode}
        onMouseLeave={this.props.leaveNode}
        onClick={this.handleMouseClick}
        x={this.props.x}
        y={this.props.y}
      />
    );
  }
}

function mapStateToProps(state) {
  return {
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
