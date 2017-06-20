import React from 'react';
import { connect } from 'react-redux';

import NodesChart from '../charts/nodes-chart';
import NodesGrid from '../charts/nodes-grid';
import NodesResources from '../components/nodes-resources';
import NodesError from '../charts/nodes-error';
import DelayedShow from '../utils/delayed-show';
import { Loading, getNodeType } from './loading';
import {
  isTopologyNodeCountZero,
  isNodesDisplayEmpty,
} from '../utils/topology-utils';
import { nodesLoadedSelector } from '../selectors/node-filters';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';

import { TOPOLOGY_LOADER_DELAY } from '../constants/timer';


// TODO: The information that we already have available on the frontend should enable
// us to determine which of these cases exactly is preventing us from seeing the nodes.
const NODES_STATS_COUNT_ZERO_CAUSES = [
  'We haven\'t received any reports from probes recently. Are the probes properly connected?',
  'Containers view only: you\'re not running Docker, or you don\'t have any containers',
];
const NODES_NOT_DISPLAYED_CAUSES = [
  'There are nodes, but they\'ve been filtered out by pinned searches in the top-left corner.',
  'There are nodes, but they\'re currently hidden. Check the view options in the bottom-left if they allow for showing hidden nodes.',
  'There are no nodes for this particular moment in time. Use the time travel feature at the bottom-right corner to explore different times.',
];

const renderCauses = causes => (
  <ul>{causes.map(cause => <li key={cause}>{cause}</li>)}</ul>
);

class Nodes extends React.Component {
  renderConditionalEmptyTopologyError() {
    const { topologyNodeCountZero, nodesDisplayEmpty } = this.props;

    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!nodesDisplayEmpty}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        {topologyNodeCountZero ?
          renderCauses(NODES_STATS_COUNT_ZERO_CAUSES) :
          renderCauses(NODES_NOT_DISPLAYED_CAUSES)}
      </NodesError>
    );
  }

  render() {
    const { topologiesLoaded, nodesLoaded, topologies, currentTopology, isGraphViewMode,
      isTableViewMode, isResourceViewMode } = this.props;

    // TODO: Rename view mode components.
    return (
      <div className="nodes-wrapper">
        <DelayedShow delay={TOPOLOGY_LOADER_DELAY} show={!topologiesLoaded || !nodesLoaded}>
          <Loading itemType="topologies" show={!topologiesLoaded} />
          <Loading
            itemType={getNodeType(currentTopology, topologies)}
            show={topologiesLoaded && !nodesLoaded} />
        </DelayedShow>

        {topologiesLoaded && nodesLoaded && this.renderConditionalEmptyTopologyError()}

        {isGraphViewMode && <NodesChart />}
        {isTableViewMode && <NodesGrid />}
        {isResourceViewMode && <NodesResources />}
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    isGraphViewMode: isGraphViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
    topologyNodeCountZero: isTopologyNodeCountZero(state),
    nodesDisplayEmpty: isNodesDisplayEmpty(state),
    nodesLoaded: nodesLoadedSelector(state),
    timeTravelTransitioning: state.get('timeTravelTransitioning'),
    currentTopology: state.get('currentTopology'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
  };
}


export default connect(
  mapStateToProps
)(Nodes);
