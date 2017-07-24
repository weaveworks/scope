import React from 'react';
// import Slider from 'rc-slider';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce, map } from 'lodash';

import TimeTravelTimeline from './time-travel-timeline';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import {
  jumpToTime,
  resumeTime,
  timeTravelStartTransition,
} from '../actions/app-actions';

import {
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';


const getTimestampStates = (timestamp) => {
  timestamp = timestamp || moment();
  return {
    inputValue: moment(timestamp).utc().format(),
  };
};

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = getTimestampStates(props.pausedAt);

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleTimelinePan = this.handleTimelinePan.bind(this);
    this.handleJumpClick = this.handleJumpClick.bind(this);
    this.travelTo = this.travelTo.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
    this.debouncedTrackSliderChange = debounce(
      this.trackSliderChange.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentWillReceiveProps(props) {
    this.setState(getTimestampStates(props.pausedAt));
  }

  componentWillUnmount() {
    // TODO: Causing bug?
    this.props.resumeTime();
  }

  handleSliderChange(timestamp) {
    if (!timestamp.isSame(this.props.pausedAt)) {
      this.travelTo(timestamp, true);
      this.debouncedTrackSliderChange();
    }
  }

  handleInputChange(ev) {
    let timestamp = moment(ev.target.value);
    this.setState({ inputValue: ev.target.value });

    if (timestamp.isValid()) {
      timestamp = Math.min(timestamp, moment().valueOf());
      this.travelTo(timestamp);

      trackMixpanelEvent('scope.time.timestamp.edit', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
    }
  }

  // TODO: Redo
  handleJumpClick(millisecondsDelta) {
    let timestamp = this.state.sliderValue + millisecondsDelta;
    timestamp = Math.min(timestamp, moment().valueOf());
    this.travelTo(timestamp, true);
  }

  updateTimestamp(timestamp) {
    this.props.jumpToTime(moment(timestamp));
  }

  travelTo(timestamp, debounced = false) {
    this.props.timeTravelStartTransition();
    this.setState(getTimestampStates(timestamp));
    if (debounced) {
      this.debouncedUpdateTimestamp(timestamp);
    } else {
      this.debouncedUpdateTimestamp.cancel();
      this.updateTimestamp(timestamp);
    }
  }

  trackSliderChange() {
    trackMixpanelEvent('scope.time.slider.change', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  render() {
    return (
      <div className={classNames('time-travel', { visible: this.props.showingTimeTravel })}>
        <TimeTravelTimeline
          onTimelinePan={this.handleTimelinePan}
          onJumpClick={this.handleJumpClick}
        />
        <div className="time-travel-timestamp">
          <input
            value={this.state.inputValue}
            onChange={this.handleInputChange}
          /> UTC
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    showingTimeTravel: state.get('showingTimeTravel'),
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
    pausedAt: state.get('pausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    jumpToTime,
    resumeTime,
    timeTravelStartTransition,
  }
)(TimeTravel);
