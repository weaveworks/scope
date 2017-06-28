import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

// import CloudFeature from './cloud-feature';
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
  }

  handleNowClick() {
    trackMixpanelEvent('scope.time.resume.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      nodesDeltaBufferSize: this.props.updateCount,
    });
    this.props.clickResumeUpdate();
  }

  handlePauseClick() {
    trackMixpanelEvent('scope.time.pause.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
    this.props.clickPauseUpdate();
  }

  handleTravelClick() {
    trackMixpanelEvent('scope.time.travel.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
    this.props.clickTimeTravel();
  }

  render() {
    const { showingTimeTravel, pausedAt, timeTravelTransitioning } = this.props;
    const isPausedNow = pausedAt && !showingTimeTravel;
    const isTimeTravelling = showingTimeTravel;
    const isRunningNow = !pausedAt;

    return (
      <div className="time-control">
        <div className="time-control-icon">
          {timeTravelTransitioning && <span className="fa fa-circle-o-notch fa-spin" />}
        </div>
        <div className="time-control-wrapper">
          <span
            className={className(isRunningNow)}
            onClick={this.handleNowClick}>
            {isRunningNow && <span className="fa fa-clock-o" />}
            <span className="label">Now</span>
          </span>
          <span
            className={className(isPausedNow)}
            onClick={this.handlePauseClick}
            disabled={isTimeTravelling}>
            {isPausedNow && <span className="fa fa-pause" />}
            <span className="label">{isPausedNow ? 'Paused' : 'Pause'}</span>
          </span>
          <span
            className={className(isTimeTravelling)}
            onClick={this.handleTravelClick}>
            <span className="label">Time Travel</span>
          </span>
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  // const cloudInstance = root.instances[params.orgId] || {};
  // const featureFlags = cloudInstance.featureFlags || [];
  return {
    // hasTimeTravel: featureFlags.includes('time-travel'),
    update: state.get('topologyViewMode'),
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
