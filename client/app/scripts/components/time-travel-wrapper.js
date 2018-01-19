import React from 'react';
import moment from 'moment';
import styled from 'styled-components';
import { connect } from 'react-redux';
import { TimeTravel } from 'weaveworks-ui-components';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { jumpToTime } from '../actions/app-actions';


const TimeTravelContainer = styled.div`
  transition: all .15s ease-in-out;
  position: relative;
  overflow: hidden;
  height: 0;

  ${props => props.visible && `
    height: 105px;
  `}
`;

class TimeTravelWrapper extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.trackTimestampEdit = this.trackTimestampEdit.bind(this);
    this.trackTimelinePanButtonClick = this.trackTimelinePanButtonClick.bind(this);
    this.trackTimelineLabelClick = this.trackTimelineLabelClick.bind(this);
    this.trackTimelineZoom = this.trackTimelineZoom.bind(this);
    this.trackTimelinePan = this.trackTimelinePan.bind(this);
  }

  trackTimestampEdit() {
    trackAnalyticsEvent('scope.time.timestamp.edit', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelinePanButtonClick() {
    trackAnalyticsEvent('scope.time.timeline.pan.button.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelineLabelClick() {
    trackAnalyticsEvent('scope.time.timeline.label.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelinePan() {
    trackAnalyticsEvent('scope.time.timeline.pan', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelineZoom(zoomedPeriod) {
    trackAnalyticsEvent('scope.time.timeline.zoom', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      zoomedPeriod,
    });
  }

  render() {
    return (
      <TimeTravelContainer visible={this.props.visible}>
        <TimeTravel
          timestamp={this.props.timestamp}
          earliestTimestamp={this.props.earliestTimestamp}
          onChangeTimestamp={this.props.jumpToTime}
          onTimestampInputEdit={this.trackTimestampEdit}
          onTimelinePanButtonClick={this.trackTimelinePanButtonClick}
          onTimelineLabelClick={this.trackTimelineLabelClick}
          onTimelineZoom={this.trackTimelineZoom}
          onTimelinePan={this.trackTimelinePan}
        />
      </TimeTravelContainer>
    );
  }
}

function mapStateToProps(state, { params }) {
  const scopeState = state.scope || state;
  let firstSeenConnectedAt;

  // If we're in the Weave Cloud context, use firstSeeConnectedAt as the earliest timestamp.
  if (state.root && state.root.instances) {
    const serviceInstance = state.root.instances[params && params.orgId];
    if (serviceInstance && serviceInstance.firstSeenConnectedAt) {
      firstSeenConnectedAt = moment(serviceInstance.firstSeenConnectedAt).utc().format();
    }
  }

  return {
    visible: scopeState.get('showingTimeTravel'),
    topologyViewMode: scopeState.get('topologyViewMode'),
    currentTopology: scopeState.get('currentTopology'),
    earliestTimestamp: firstSeenConnectedAt,
    timestamp: scopeState.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  { jumpToTime },
)(TimeTravelWrapper);
