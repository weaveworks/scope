import React from 'react';
import Slider from 'rc-slider';
import moment from 'moment';
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


const ONE_HOUR_MS = moment.duration(1, 'hour');
const FIVE_MINUTES_MS = moment.duration(5, 'minutes');

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      // TODO: Showing a three months of history is quite arbitrary;
      // we should instead get some meaningful 'beginning of time' from
      // the backend and make the slider show whole active history.
      sliderMinValue: moment().subtract(6, 'months').valueOf(),
      sliderValue: props.pausedAt && props.pausedAt.valueOf(),
      inputValue: props.pausedAt && moment(props.pausedAt).utc().format(),
    };

    this.handleTimestampInputChange = this.handleTimestampInputChange.bind(this);
    this.handleTimestampClick = this.handleTimestampClick.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);
    this.jumpInTime = this.jumpInTime.bind(this);
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
    this.setState({
      sliderValue: props.pausedAt && props.pausedAt.valueOf(),
      inputValue: props.pausedAt && moment(props.pausedAt).utc().format(),
    });
  }

  componentWillUnmount() {
    clearInterval(this.timer);
    this.props.clickResumeUpdate();
  }

  handleSliderChange(value) {
    const timestamp = moment(value).utc();

    this.setState({
      inputValue: timestamp.format(),
      sliderValue: value,
    });

    this.props.timeTravelStartTransition();
    this.debouncedUpdateTimestamp(timestamp);
    this.debouncedTrackSliderChange();
  }

  handleTimestampClick() {
    trackMixpanelEvent('scope.time.timestamp.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  handleTimestampInputChange(ev) {
    this.setState({ inputValue: ev.target.value });
  }

  updateTimestamp(timestamp) {
    this.props.jumpToTime(moment(timestamp));
  }

  jumpInTime(millisecondsDelta) {
    let timestamp = this.state.sliderValue - millisecondsDelta;
    timestamp = Math.min(timestamp, this.state.sliderMinValue);
    timestamp = Math.max(timestamp, moment().valueOf());

    this.props.timeTravelStartTransition();
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
    const style = { marginLeft: `calc(${(1 - pos) * 100}% - 50px)`, width: '100px' };
    return (
      <div
        style={style}
        className="time-travel-markers-tick"
        key={timestampValue}>
        <span className="vertical-tick" />
        <a className="link" onClick={() => this.updateTimestamp(timestampValue)}>{label}</a>
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
    // Don't render the time travel control if it's not explicitly enabled for this instance.
    if (!this.props.showingTimeTravel) return null;

    const { sliderValue, sliderMinValue, inputValue } = this.state;
    const sliderMaxValue = moment().valueOf();

    return (
      <div className="time-travel">
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
          <a className="button jump" onClick={() => this.jumpInTime(-ONE_HOUR_MS)}>
            <span className="fa fa-fast-backward" /> 1 hour
          </a>
          <a className="button jump" onClick={() => this.jumpInTime(-FIVE_MINUTES_MS)}>
            <span className="fa fa-step-backward" /> 5 mins
          </a>
          <span className="time-travel-jump-controls-timestamp">
            <input value={inputValue} onChange={this.handleTimestampInputChange} /> UTC
          </span>
          <a className="button jump" onClick={() => this.jumpInTime(+FIVE_MINUTES_MS)}>
            <span className="fa fa-step-forward" /> 5 mins
          </a>
          <a className="button jump" onClick={() => this.jumpInTime(+ONE_HOUR_MS)}>
            <span className="fa fa-fast-forward" /> 1 hour
          </a>
        </div>
      </div>
    );
  }
}

function mapStateToProps({ scope, root }, { params }) {
  const cloudInstance = root.instances[params.orgId] || {};
  const featureFlags = cloudInstance.featureFlags || [];
  return {
    showingTimeTravel: featureFlags.includes('time-travel') && scope.get('showingTimeTravel'),
    topologyViewMode: scope.get('topologyViewMode'),
    currentTopology: scope.get('currentTopology'),
    pausedAt: scope.get('pausedAt'),
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
