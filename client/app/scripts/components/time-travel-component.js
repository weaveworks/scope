import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { debounce } from 'lodash';

import TimeTravelTimeline from './time-travel-timeline';
import { clampToNowInSecondsPrecision } from '../utils/time-utils';

import { TIMELINE_DEBOUNCE_INTERVAL } from '../constants/timer';


const getTimestampStates = (timestamp) => {
  timestamp = timestamp || moment();
  return {
    inputValue: moment(timestamp).utc().format(),
  };
};

export default class TimeTravelComponent extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = getTimestampStates(props.timestamp);

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleTimelinePan = this.handleTimelinePan.bind(this);
    this.handleTimelinePanEnd = this.handleTimelinePanEnd.bind(this);
    this.handleInstantJump = this.handleInstantJump.bind(this);

    this.instantUpdateTimestamp = this.instantUpdateTimestamp.bind(this);
    this.debouncedUpdateTimestamp = debounce(
      this.instantUpdateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentWillReceiveProps(props) {
    this.setState(getTimestampStates(props.timestamp));
  }

  handleInputChange(ev) {
    const timestamp = moment(ev.target.value);
    this.setState({ inputValue: ev.target.value });

    if (timestamp.isValid()) {
      const clampedTimestamp = clampToNowInSecondsPrecision(timestamp);
      this.instantUpdateTimestamp(clampedTimestamp, this.props.trackTimestampEdit);
    }
  }

  handleTimelinePan(timestamp) {
    this.setState(getTimestampStates(timestamp));
    this.debouncedUpdateTimestamp(timestamp);
  }

  handleTimelinePanEnd(timestamp) {
    this.instantUpdateTimestamp(timestamp, this.props.trackTimelinePan);
  }

  handleInstantJump(timestamp) {
    this.instantUpdateTimestamp(timestamp, this.props.trackTimelineClick);
  }

  instantUpdateTimestamp(timestamp, callback) {
    if (!timestamp.isSame(this.props.timestamp)) {
      this.debouncedUpdateTimestamp.cancel();
      this.setState(getTimestampStates(timestamp));
      this.props.changeTimestamp(moment(timestamp));

      // Used for tracking.
      if (callback) callback();
    }
  }

  render() {
    const { visible, timestamp, viewportWidth } = this.props;

    return (
      <div className={classNames('time-travel', { visible })}>
        <TimeTravelTimeline
          timestamp={timestamp}
          viewportWidth={viewportWidth}
          onTimelinePan={this.handleTimelinePan}
          onTimelinePanEnd={this.handleTimelinePanEnd}
          onInstantJump={this.handleInstantJump}
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
