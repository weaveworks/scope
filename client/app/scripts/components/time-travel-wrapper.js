import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';
import { TimeTravel } from 'weaveworks-ui-components';
import { get, orderBy, debounce } from 'lodash';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { jumpToTime, resumeTime, pauseTimeAtNow, getFluxHistory } from '../actions/app-actions';

// Load deployments in timeline only on zoom levels up to this range.
const MAX_DEPLOYMENTS_RANGE_SECS = moment.duration(2, 'weeks').asSeconds();

// Reused from Service UI.
const FLUX_ALL_SERVICES = '<all>';

class TimeTravelWrapper extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      isLoadingDeployments: false,
      visibleRangeStartAtSec: null,
      visibleRangeEndAtSec: null,
    };

    this.debouncedUpdateVisibleRange = debounce(this.updateVisibleRange, 500);
  }

  trackTimestampEdit = () => {
    trackAnalyticsEvent('scope.time.timestamp.edit', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelinePanButtonClick = () => {
    trackAnalyticsEvent('scope.time.timeline.pan.button.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelineLabelClick = () => {
    trackAnalyticsEvent('scope.time.timeline.label.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelinePan = () => {
    trackAnalyticsEvent('scope.time.timeline.pan', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelineZoom = (zoomedPeriod) => {
    trackAnalyticsEvent('scope.time.timeline.zoom', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      zoomedPeriod,
    });
  }

  handleLiveModeChange = (showingLive) => {
    if (showingLive) {
      this.props.resumeTime();
    } else {
      this.props.pauseTimeAtNow();
    }
  }


  updateVisibleRange = ({ startAt, endAt }) => {
    const { orgId } = this.props.params;

    const visibleRangeEndAtSec = moment(endAt).unix();
    const visibleRangeStartAtSec = moment(startAt).unix();
    this.setState({ visibleRangeStartAtSec, visibleRangeEndAtSec });

    // Load deployment annotations only if not zoomed out too much.
    // See https://github.com/weaveworks/service-ui/issues/1858.
    const visibleRangeSec = visibleRangeEndAtSec - visibleRangeStartAtSec;
    if (visibleRangeSec < MAX_DEPLOYMENTS_RANGE_SECS) {
      this.setState({ isLoadingDeployments: true });
      this.props
        .getFluxHistory(orgId, FLUX_ALL_SERVICES, null, endAt, true, startAt)
        .then(() => {
          this.setState({ isLoadingDeployments: false });
        });
    }
  }

  render() {
    const {
      visibleRangeStartAtSec,
      visibleRangeEndAtSec,
    } = this.state;

    // Don't pass any deployments that are outside of the timeline visible range.
    const visibleDeployments = this.props.deployments.filter((deployment) => {
      const deploymentAtSec = moment(deployment.Stamp).unix();
      return (
        visibleRangeStartAtSec <= deploymentAtSec &&
        deploymentAtSec <= visibleRangeEndAtSec
      );
    });

    return (
      <div className="tour-step-anchor time-travel-wrapper">
        <TimeTravel
          hasLiveMode
          showingLive={this.props.showingLive}
          onChangeLiveMode={this.handleLiveModeChange}
          timestamp={this.props.timestamp}
          earliestTimestamp={this.props.earliestTimestamp}
          onChangeTimestamp={this.props.jumpToTime}
          deployments={visibleDeployments}
          isLoading={this.state.isLoadingDeployments}
          onUpdateVisibleRange={this.debouncedUpdateVisibleRange}
          onTimestampInputEdit={this.trackTimestampEdit}
          onTimelinePanButtonClick={this.trackTimelinePanButtonClick}
          onTimelineLabelClick={this.trackTimelineLabelClick}
          onTimelineZoom={this.trackTimelineZoom}
          onTimelinePan={this.trackTimelinePan}
        />
      </div>
    );
  }
}

function mapStateToProps(state, { params }) {
  const orgId = params && params.orgId;
  let firstSeenConnectedAt;

  // If we're in the Weave Cloud context, use firstSeeConnectedAt as the earliest timestamp.
  if (state.root && state.root.instances) {
    const serviceInstance = state.root.instances[orgId];
    if (serviceInstance && serviceInstance.firstSeenConnectedAt) {
      firstSeenConnectedAt = moment(serviceInstance.firstSeenConnectedAt).utc().format();
    }
  }

  const unsortedDeployments = get(
    state,
    ['root', 'fluxInstanceHistory', orgId, FLUX_ALL_SERVICES],
    []
  );

  return {
    showingLive: !state.scope.get('pausedAt'),
    topologyViewMode: state.scope.get('topologyViewMode'),
    currentTopology: state.scope.get('currentTopology'),
    deployments: orderBy(unsortedDeployments, ['Stamp'], ['desc']),
    earliestTimestamp: firstSeenConnectedAt,
    timestamp: state.scope.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    jumpToTime, resumeTime, pauseTimeAtNow, getFluxHistory
  },
)(TimeTravelWrapper);
