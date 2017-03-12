import React from 'react';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import NodesChart from '../charts/nodes-chart';
import NodesGrid from '../charts/nodes-grid';
import NodesResources from '../charts/nodes-resources';
import NodesError from '../charts/nodes-error';
import DelayedShow from '../utils/delayed-show';
import { Loading, getNodeType } from './loading';
import { isTopologyEmpty } from '../utils/topology-utils';
import { setViewportDimensions } from '../actions/app-actions';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';

import { VIEWPORT_RESIZE_DEBOUNCE_INTERVAL } from '../constants/timer';


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
  constructor(props, context) {
    super(props, context);

    this.setDimensions = this.setDimensions.bind(this);
    this.handleResize = debounce(this.setDimensions, VIEWPORT_RESIZE_DEBOUNCE_INTERVAL);
    this.setDimensions();
  }

  componentDidMount() {
    window.addEventListener('resize', this.handleResize);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
  }

  render() {
    const { topologyEmpty, topologiesLoaded, nodesLoaded, topologies, currentTopology,
      isGraphViewMode, isTableViewMode, isResourceViewMode } = this.props;

    // TODO: Rename view mode components.
    return (
      <div className="nodes-wrapper">
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

  setDimensions() {
    this.props.setViewportDimensions(window.innerWidth, window.innerHeight);
  }
}


function mapStateToProps(state) {
  return {
    isGraphViewMode: isGraphViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
    currentTopology: state.get('currentTopology'),
    nodesLoaded: state.get('nodesLoaded'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topologyEmpty: isTopologyEmpty(state),
  };
}


export default connect(
  mapStateToProps,
  { setViewportDimensions }
)(Nodes);
