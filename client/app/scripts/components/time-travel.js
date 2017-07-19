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
  TIMELINE_TICK_INTERVAL,
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';


const getTimestampStates = (timestamp) => {
  timestamp = timestamp || moment();
  return {
    sliderValue: moment(timestamp).valueOf(),
    inputValue: moment(timestamp).utc().format(),
  };
};

// const ONE_HOUR_MS = moment.duration(1, 'hour');
// const FIVE_MINUTES_MS = moment.duration(5, 'minutes');

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      // TODO: Showing a three months of history is quite arbitrary;
      // we should instead get some meaningful 'beginning of time' from
      // the backend and make the slider show whole active history.
      sliderMinValue: moment().subtract(6, 'months').valueOf(),
      ...getTimestampStates(props.pausedAt),
    };

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);
    this.handleJumpClick = this.handleJumpClick.bind(this);
    this.renderMarks = this.renderMarks.bind(this);
    this.renderMark = this.renderMark.bind(this);
    this.travelTo = this.travelTo.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
    this.debouncedTrackSliderChange = debounce(
      this.trackSliderChange.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    // Force periodic re-renders to update the slider position as time goes by.
    this.timer = setInterval(() => { this.forceUpdate(); }, TIMELINE_TICK_INTERVAL);
  }

  componentWillReceiveProps(props) {
    this.setState(getTimestampStates(props.pausedAt));
  }

  componentWillUnmount() {
    clearInterval(this.timer);
    this.props.resumeTime();
  }

  handleSliderChange(timestamp) {
    this.travelTo(timestamp, true);
    this.debouncedTrackSliderChange();
  }

  handleInputChange(ev) {
    let timestamp = moment(ev.target.value);
    this.setState({ inputValue: ev.target.value });

    if (timestamp.isValid()) {
      timestamp = Math.max(timestamp, this.state.sliderMinValue);
      timestamp = Math.min(timestamp, moment().valueOf());
      this.travelTo(timestamp, true);

      trackMixpanelEvent('scope.time.timestamp.edit', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
    }
  }

  handleJumpClick(millisecondsDelta) {
    let timestamp = this.state.sliderValue + millisecondsDelta;
    timestamp = Math.max(timestamp, this.state.sliderMinValue);
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

  renderMark({ timestampValue, label }) {
    const sliderMaxValue = moment().valueOf();
    const pos = (sliderMaxValue - timestampValue) / (sliderMaxValue - this.state.sliderMinValue);

    // Ignore the month marks that are very close to 'Now'
    if (label !== 'Now' && pos < 0.05) return null;

    const style = { marginLeft: `calc(${(1 - pos) * 100}% - 32px)`, width: '64px' };
    return (
      <div
        style={style}
        className="time-travel-markers-tick"
        key={timestampValue}>
        <span className="vertical-tick" />
        <a className="link" onClick={() => this.travelTo(timestampValue)}>{label}</a>
      </div>
    );
  }

  renderMarks() {
    const { sliderMinValue } = this.state;
    const sliderMaxValue = moment().valueOf();
    const ticks = [{ timestampValue: sliderMaxValue, label: 'Now' }];
    let monthsBack = 0;
    let timestamp;

    do {
      timestamp = moment().utc().subtract(monthsBack, 'months').startOf('month');
      if (timestamp.valueOf() >= sliderMinValue) {
        // Months are broken by the year tag, e.g. November, December, 2016, February, etc...
        let label = timestamp.format('MMMM');
        if (label === 'January') {
          label = timestamp.format('YYYY');
        }
        ticks.push({ timestampValue: timestamp.valueOf(), label });
      }
      monthsBack += 1;
    } while (timestamp.valueOf() >= sliderMinValue);

    return (
      <div className="time-travel-markers">
        {map(ticks, tick => this.renderMark(tick))}
      </div>
    );
  }

  render() {
    const { inputValue } = this.state;
    // const { sliderValue, sliderMinValue, inputValue } = this.state;
    // const sliderMaxValue = moment().valueOf();

    const className = classNames('time-travel', { visible: this.props.showingTimeTravel });

    return (
      <div className={className}>
        <TimeTravelTimeline />
        <div className="time-travel-timestamp">
          <input value={inputValue} onChange={this.handleInputChange} /> UTC
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
