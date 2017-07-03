import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import CloudFeature from './cloud-feature';
import TimeTravelButton from './time-travel-button';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import { clickPauseUpdate, clickResumeUpdate, clickTimeTravel } from '../actions/app-actions';


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

  getTrackingMetadata() {
    const { currentTopology } = this.props;
    return {
      layout: this.props.topologyViewMode,
      topologyId: currentTopology && currentTopology.get('id'),
      parentTopologyId: currentTopology && currentTopology.get('parentId'),
    };
  }

  handleNowClick() {
    trackMixpanelEvent('scope.time.resume.click', this.getTrackingMetadata());
    this.props.clickResumeUpdate();
  }

  handlePauseClick() {
    trackMixpanelEvent('scope.time.pause.click', this.getTrackingMetadata());
    this.props.clickPauseUpdate();
  }

  handleTravelClick() {
    trackMixpanelEvent('scope.time.travel.click', this.getTrackingMetadata());
    this.props.clickTimeTravel();
  }

  render() {
    const { showingTimeTravel, pausedAt, timeTravelTransitioning } = this.props;

    const isPausedNow = pausedAt && !showingTimeTravel;
    const isTimeTravelling = showingTimeTravel;
    const isRunningNow = !pausedAt;

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
            <CloudFeature>
              <TimeTravelButton
                className={className(isTimeTravelling)}
                onClick={this.handleTravelClick}
                isTimeTravelling={isTimeTravelling}
              />
            </CloudFeature>
          </div>
        </div>
        {isPausedNow && <span
          className="time-control-info"
          title={moment(pausedAt).utc().toISOString()}>
          Paused {moment(pausedAt).fromNow()}
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
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
    showingTimeTravel: state.get('showingTimeTravel'),
    timeTravelTransitioning: state.get('timeTravelTransitioning'),
    pausedAt: state.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    clickPauseUpdate,
    clickResumeUpdate,
    clickTimeTravel,
  }
)(TimeControl);
