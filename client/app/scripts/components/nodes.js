import React from 'react';
import classNames from 'classnames';
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
  isTopologyEmpty,
} from '../utils/topology-utils';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';

import { TOPOLOGY_LOADER_DELAY } from '../constants/timer';


// TODO: The information that we already have available on the frontend should enable
// us to determine which of these cases exactly is preventing us from seeing the nodes.
const NODE_COUNT_ZERO_CAUSES = [
  'We haven\'t received any reports from probes recently. Are the probes properly connected?',
  'Containers view only: you\'re not running Docker, or you don\'t have any containers',
];
const NODES_DISPLAY_EMPTY_CAUSES = [
  'There are nodes, but they\'re currently hidden. Check the view options in the bottom-left if they allow for showing hidden nodes.',
  'There are no nodes for this particular moment in time. Use the time travel feature at the bottom-right corner to explore different times.',
];

const renderCauses = causes => (
  <ul>{causes.map(cause => <li key={cause}>{cause}</li>)}</ul>
);

class Nodes extends React.Component {
  renderConditionalEmptyTopologyError() {
    const { topologyNodeCountZero, nodesDisplayEmpty, topologyEmpty } = this.props;

    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!topologyEmpty}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        {topologyNodeCountZero && renderCauses(NODE_COUNT_ZERO_CAUSES)}
        {!topologyNodeCountZero && nodesDisplayEmpty && renderCauses(NODES_DISPLAY_EMPTY_CAUSES)}
      </NodesError>
    );
  }

  render() {
    const { topologiesLoaded, nodesLoaded, topologies, currentTopology, isGraphViewMode,
      isTableViewMode, isResourceViewMode, websocketTransitioning } = this.props;

    const className = classNames('nodes-wrapper', { blurred: websocketTransitioning });

    // TODO: Rename view mode components.
    return (
      <div className={className}>
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
    topologyEmpty: isTopologyEmpty(state),
    websocketTransitioning: state.get('websocketTransitioning'),
    currentTopology: state.get('currentTopology'),
    nodesLoaded: state.get('nodesLoaded'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
  };
}


export default connect(
  mapStateToProps
)(Nodes);
