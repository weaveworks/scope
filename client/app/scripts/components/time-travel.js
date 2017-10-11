import React from 'react';
import { connect } from 'react-redux';

import TimeTravelComponent from './time-travel-component';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import {
  jumpToTime,
  timeTravelStartTransition,
} from '../actions/app-actions';

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.changeTimestamp = this.changeTimestamp.bind(this);
    this.trackTimestampEdit = this.trackTimestampEdit.bind(this);
    this.trackTimelineClick = this.trackTimelineClick.bind(this);
    this.trackTimelinePan = this.trackTimelinePan.bind(this);
  }

  changeTimestamp(timestamp) {
    this.props.timeTravelStartTransition();
    this.props.jumpToTime(timestamp);
  }

  trackTimestampEdit() {
    trackAnalyticsEvent('scope.time.timestamp.edit', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  trackTimelineClick() {
    trackAnalyticsEvent('scope.time.timeline.click', {
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

  render() {
    const { visible, timestamp, viewportWidth } = this.props;

    return (
      <TimeTravelComponent
        visible={visible}
        timestamp={timestamp}
        viewportWidth={viewportWidth}
        changeTimestamp={this.changeTimestamp}
        trackTimestampEdit={this.trackTimestampEdit}
        trackTimelineClick={this.trackTimelineClick}
        trackTimelinePan={this.trackTimelinePan}
      />
    );
  }
}

function mapStateToProps(state) {
  return {
    visible: state.get('showingTimeTravel'),
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
    timestamp: state.get('pausedAt'),
    // Used only to trigger recalculations on window resize.
    viewportWidth: state.getIn(['viewport', 'width']),
  };
}

export default connect(
  mapStateToProps,
  { jumpToTime, timeTravelStartTransition },
)(TimeTravel);
