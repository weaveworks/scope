import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';
import { TimeTravel } from 'weaveworks-ui-components';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { jumpToTime } from '../actions/app-actions';


class TimeTravelWrapper extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.changeTimestamp = this.changeTimestamp.bind(this);
    this.trackTimestampEdit = this.trackTimestampEdit.bind(this);
    this.trackTimelineClick = this.trackTimelineClick.bind(this);
    this.trackTimelineZoom = this.trackTimelineZoom.bind(this);
    this.trackTimelinePan = this.trackTimelinePan.bind(this);
  }

  changeTimestamp(timestamp) {
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

  trackTimelineZoom(zoomedPeriod) {
    trackAnalyticsEvent('scope.time.timeline.zoom', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      zoomedPeriod,
    });
  }

  render() {
    const { visible, timestamp } = this.props;

    return (
      <TimeTravel
        visible={visible}
        timestamp={timestamp || moment()}
        onChange={this.changeTimestamp}
        onTimestampInputEdit={this.trackTimestampEdit}
        onTimestampLabelClick={this.trackTimelineClick}
        onTimelineZoom={this.trackTimelineZoom}
        onTimelinePan={this.trackTimelinePan}
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
  };
}

export default connect(
  mapStateToProps,
  { jumpToTime },
)(TimeTravelWrapper);
