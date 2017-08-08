import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { trackMixpanelEvent } from '../utils/tracking-utils';
import { pauseTimeAtNow, resumeTime, startTimeTravel } from '../actions/app-actions';

import { TIMELINE_TICK_INTERVAL } from '../constants/timer';


const className = isSelected => (
  classNames('time-control-action', { 'time-control-action-selected': isSelected })
);

class TimeControl extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleNowClick = this.handleNowClick.bind(this);
    this.handlePauseClick = this.handlePauseClick.bind(this);
    this.handleTravelClick = this.handleTravelClick.bind(this);
    this.getTrackingMetadata = this.getTrackingMetadata.bind(this);
  }

  componentDidMount() {
    // Force periodic for the paused info.
    this.timer = setInterval(() => { this.forceUpdate(); }, TIMELINE_TICK_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  getTrackingMetadata(data = {}) {
    const { currentTopology } = this.props;
    return {
      layout: this.props.topologyViewMode,
      topologyId: currentTopology && currentTopology.get('id'),
      parentTopologyId: currentTopology && currentTopology.get('parentId'),
      ...data
    };
  }

  handleNowClick() {
    trackMixpanelEvent('scope.time.resume.click', this.getTrackingMetadata());
    this.props.resumeTime();
  }

  handlePauseClick() {
    trackMixpanelEvent('scope.time.pause.click', this.getTrackingMetadata());
    this.props.pauseTimeAtNow();
  }

  handleTravelClick() {
    if (!this.props.showingTimeTravel) {
      trackMixpanelEvent('scope.time.travel.click', this.getTrackingMetadata({ open: true }));
      this.props.startTimeTravel();
    } else {
      trackMixpanelEvent('scope.time.travel.click', this.getTrackingMetadata({ open: false }));
      this.props.resumeTime();
    }
  }

  render() {
    const { showingTimeTravel, pausedAt, timeTravelTransitioning, topologiesLoaded,
      hasHistoricReports } = this.props;

    const isPausedNow = pausedAt && !showingTimeTravel;
    const isTimeTravelling = showingTimeTravel;
    const isRunningNow = !pausedAt;

    if (!topologiesLoaded) return null;

    return (
      <div className="time-control">
        <div className="time-control-controls">
          <div className="time-control-spinner">
            {timeTravelTransitioning && <span className="fa fa-circle-o-notch fa-spin" />}
          </div>
          <div className="time-control-wrapper">
            <span
              className={className(isRunningNow)}
              onClick={this.handleNowClick}
              title="Show live state of the system">
              {isRunningNow && <span className="fa fa-play" />}
              <span className="label">Live</span>
            </span>
            <span
              className={className(isPausedNow)}
              onClick={!isTimeTravelling && this.handlePauseClick}
              disabled={isTimeTravelling}
              title="Pause updates (freezes the nodes in their current layout)">
              {isPausedNow && <span className="fa fa-pause" />}
              <span className="label">{isPausedNow ? 'Paused' : 'Pause'}</span>
            </span>
            {hasHistoricReports && <span
              className={className(isTimeTravelling)}
              onClick={this.handleTravelClick}
              title="Travel back in time">
              {isTimeTravelling && <span className="fa fa-clock-o" />}
              <span className="label">Time Travel</span>
            </span>}
          </div>
        </div>
        {(isPausedNow || isTimeTravelling) && <span
          className="time-control-info"
          title={moment(pausedAt).toISOString()}>
          Showing state from {moment(pausedAt).fromNow()}
        </span>}
        {isRunningNow && timeTravelTransitioning && <span
          className="time-control-info">
          Resuming the live state
        </span>}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    hasHistoricReports: state.getIn(['capabilities', 'historic_reports']),
    topologyViewMode: state.get('topologyViewMode'),
    topologiesLoaded: state.get('topologiesLoaded'),
    currentTopology: state.get('currentTopology'),
    showingTimeTravel: state.get('showingTimeTravel'),
    timeTravelTransitioning: state.get('timeTravelTransitioning'),
    pausedAt: state.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    resumeTime,
    pauseTimeAtNow,
    startTimeTravel,
  }
)(TimeControl);
