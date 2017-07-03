import React from 'react';
import Slider from 'rc-slider';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce, map } from 'lodash';

import { trackMixpanelEvent } from '../utils/tracking-utils';
import {
  jumpToTime,
  clickResumeUpdate,
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

const ONE_HOUR_MS = moment.duration(1, 'hour');
const FIVE_MINUTES_MS = moment.duration(5, 'minutes');

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      // TODO: Showing a three months of history is quite arbitrary;
      // we should instead get some meaningful 'beginning of time' from
      // the backend and make the slider show whole active history.
      sliderMinValue: moment().subtract(15, 'months').valueOf(),
      ...getTimestampStates(props.pausedAt),
    };

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);
    this.handleJumpClick = this.handleJumpClick.bind(this);
    this.debouncedJumpTo = this.debouncedJumpTo.bind(this);
    this.renderMarks = this.renderMarks.bind(this);
    this.renderMark = this.renderMark.bind(this);

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
    this.props.clickResumeUpdate();
  }

  handleSliderChange(timestamp) {
    this.debouncedJumpTo(timestamp);
    this.debouncedTrackSliderChange();
  }

  handleInputChange(ev) {
    let timestamp = moment(ev.target.value);
    this.setState({ inputValue: ev.target.value });

    if (timestamp.isValid()) {
      timestamp = Math.max(timestamp, this.state.sliderMinValue);
      timestamp = Math.min(timestamp, moment().valueOf());
      this.debouncedJumpTo(timestamp);

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
    this.debouncedJumpTo(timestamp);
  }

  updateTimestamp(timestamp) {
    this.props.jumpToTime(moment(timestamp));
  }

  debouncedJumpTo(timestamp) {
    this.props.timeTravelStartTransition();
    this.setState(getTimestampStates(timestamp));
    this.debouncedUpdateTimestamp(timestamp);
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
    if (label !== 'Now' && pos < 0.05) return null;

    const style = { marginLeft: `calc(${(1 - pos) * 100}% - 32px)`, width: '64px' };
    return (
      <div
        style={style}
        className="time-travel-markers-tick"
        key={timestampValue}>
        <span className="vertical-tick" />
        <a className="link" onClick={() => this.debouncedJumpTo(timestampValue)}>{label}</a>
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
    const { sliderValue, sliderMinValue, inputValue } = this.state;
    const sliderMaxValue = moment().valueOf();

    const className = classNames('time-travel', { visible: this.props.showingTimeTravel });

    return (
      <div className={className}>
        <div className="time-travel-slider-wrapper">
          {this.renderMarks()}
          <Slider
            onChange={this.handleSliderChange}
            value={sliderValue}
            min={sliderMinValue}
            max={sliderMaxValue}
          />
        </div>
        <div className="time-travel-jump-controls">
          <a className="button jump" onClick={() => this.handleJumpClick(-ONE_HOUR_MS)}>
            <span className="fa fa-fast-backward" /> 1 hour
          </a>
          <a className="button jump" onClick={() => this.handleJumpClick(-FIVE_MINUTES_MS)}>
            <span className="fa fa-step-backward" /> 5 mins
          </a>
          <span className="time-travel-jump-controls-timestamp">
            <input value={inputValue} onChange={this.handleInputChange} /> UTC
          </span>
          <a className="button jump" onClick={() => this.handleJumpClick(+FIVE_MINUTES_MS)}>
            <span className="fa fa-step-forward" /> 5 mins
          </a>
          <a className="button jump" onClick={() => this.handleJumpClick(+ONE_HOUR_MS)}>
            <span className="fa fa-fast-forward" /> 1 hour
          </a>
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
    clickResumeUpdate,
    timeTravelStartTransition,
  }
)(TimeTravel);
