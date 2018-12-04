import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { TimestampTag } from 'weaveworks-ui-components';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { pauseTimeAtNow, resumeTime } from '../actions/app-actions';
import { isPausedSelector, timeTravelSupportedSelector } from '../selectors/time-travel';


const className = isSelected => (
  classNames('time-control-action', { 'time-control-action-selected': isSelected })
);

class TimeControl extends React.Component {
  getTrackingMetadata = (data = {}) => {
    const { currentTopology } = this.props;
    return {
      layout: this.props.topologyViewMode,
      topologyId: currentTopology && currentTopology.get('id'),
      parentTopologyId: currentTopology && currentTopology.get('parentId'),
      ...data
    };
  }

  handleNowClick = () => {
    trackAnalyticsEvent('scope.time.resume.click', this.getTrackingMetadata());
    this.props.resumeTime();
  }

  handlePauseClick = () => {
    trackAnalyticsEvent('scope.time.pause.click', this.getTrackingMetadata());
    this.props.pauseTimeAtNow();
  }

  render() {
    const { isPaused, pausedAt, topologiesLoaded } = this.props;

    // If Time Travel is supported, show an empty placeholder div instead
    // of this control, since time will be controlled through the timeline.
    // We return <div /> instead of null so that selector controls would
    // be aligned the same way between WC Explore and Scope standalone.
    if (this.props.timeTravelSupported) return <div />;

    return (
      <div className="time-control">
        <div className="time-control-controls">
          <div className="time-control-wrapper">
            <span
              className={className(!isPaused)}
              onClick={this.handleNowClick}
              disabled={!topologiesLoaded}
              title="Show live state of the system">
              {!isPaused && <span className="fa fa-play" />}
              <span className="label">Live</span>
            </span>
            <span
              className={className(isPaused)}
              onClick={this.handlePauseClick}
              disabled={!topologiesLoaded}
              title="Pause updates (freezes the nodes in their current layout)">
              {isPaused && <span className="fa fa-pause" />}
              <span className="label">{isPaused ? 'Paused' : 'Pause'}</span>
            </span>
          </div>
        </div>
        {isPaused &&
          <span className="time-control-info">
            Showing state from <TimestampTag inheritStyles relative timestamp={pausedAt} />
          </span>
        }
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    isPaused: isPausedSelector(state),
    timeTravelSupported: timeTravelSupportedSelector(state),
    topologyViewMode: state.get('topologyViewMode'),
    topologiesLoaded: state.get('topologiesLoaded'),
    currentTopology: state.get('currentTopology'),
    pausedAt: state.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    resumeTime,
    pauseTimeAtNow,
  }
)(TimeControl);
