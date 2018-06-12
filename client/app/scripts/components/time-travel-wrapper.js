import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';
import { TimeTravel } from 'weaveworks-ui-components';
import { get, orderBy } from 'lodash';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { jumpToTime, resumeTime, pauseTimeAtNow } from '../actions/app-actions';


class TimeTravelWrapper extends React.Component {
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

  render() {
    return (
      <div className="tour-step-anchor time-travel-wrapper">
        <TimeTravel
          hasLiveMode
          showingLive={this.props.showingLive}
          onChangeLiveMode={this.handleLiveModeChange}
          timestamp={this.props.timestamp}
          earliestTimestamp={this.props.earliestTimestamp}
          onChangeTimestamp={this.props.jumpToTime}
          deployments={this.props.deployments}
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

  const unsortedDeployments = get(state, ['root', 'fluxInstanceHistory', orgId, '<all>'], []);

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
  { jumpToTime, resumeTime, pauseTimeAtNow },
)(TimeTravelWrapper);
