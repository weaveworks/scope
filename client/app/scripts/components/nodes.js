import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import NodesChart from '../charts/nodes-chart';
import NodesGrid from '../charts/nodes-grid';
import NodesResources from '../components/nodes-resources';
import NodesError from '../charts/nodes-error';
import DelayedShow from '../utils/delayed-show';
import { Loading, getNodeType } from './loading';
import { isTopologyEmpty } from '../utils/topology-utils';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';


const EmptyTopologyError = show => (
  <NodesError faIconClass="fa-circle-thin" hidden={!show}>
    <div className="heading">Nothing to show. This can have any of these reasons:</div>
    <ul>
      <li>We haven&apos;t received any reports from probes recently.
       Are the probes properly configured?</li>
      <li>There are nodes, but they&apos;re currently hidden. Check the view options
       in the bottom-left if they allow for showing hidden nodes.</li>
      <li>Containers view only: you&apos;re not running Docker,
       or you don&apos;t have any containers.</li>
    </ul>
  </NodesError>
);

class Nodes extends React.Component {
  render() {
    const { topologyEmpty, topologiesLoaded, nodesLoaded, topologies, currentTopology,
      isGraphViewMode, isTableViewMode, isResourceViewMode, blurred } = this.props;

    const className = classNames('nodes-wrapper', { blurred });

    // TODO: Rename view mode components.
    return (
      <div className={className}>
        <DelayedShow delay={1000} show={!topologiesLoaded || (topologiesLoaded && !nodesLoaded)}>
          <Loading itemType="topologies" show={!topologiesLoaded} />
          <Loading
            itemType={getNodeType(currentTopology, topologies)}
            show={topologiesLoaded && !nodesLoaded} />
        </DelayedShow>
        {EmptyTopologyError(topologiesLoaded && nodesLoaded && topologyEmpty)}

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
    blurred: state.get('websocketMovingInTime'),
    currentTopology: state.get('currentTopology'),
    nodesLoaded: state.get('nodesLoaded'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topologyEmpty: isTopologyEmpty(state),
  };
}


export default connect(
  mapStateToProps
)(Nodes);
