import React from 'react';
import Slider from 'rc-slider';
import moment from 'moment';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

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


const ONE_HOUR_MS = 60 * 60 * 1000;
const FIVE_MINUTES_MS = 5 * 60 * 1000;

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      // TODO: Showing a three months of history is quite arbitrary;
      // we should instead get some meaningful 'beginning of time' from
      // the backend and make the slider show whole active history.
      sliderMinValue: moment().subtract(3, 'months').valueOf(),
      sliderValue: props.pausedAt && props.pausedAt.valueOf(),
      inputValue: props.pausedAt && moment(props.pausedAt).utc().format(),
    };

    this.handleTimestampInputChange = this.handleTimestampInputChange.bind(this);
    this.handleTimestampClick = this.handleTimestampClick.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);
    this.jumpInTime = this.jumpInTime.bind(this);

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
    this.props.jumpToTime(timestamp);
  }

  jumpInTime(millisecondsDelta) {
    let timestamp = this.state.sliderValue - millisecondsDelta;
    timestamp = Math.min(timestamp, this.state.sliderStartTimestamp);
    timestamp = Math.max(timestamp, moment().valueOf());

    this.props.timeTravelStartTransition();
    this.debouncedUpdateTimestamp(moment(timestamp));
  }

  trackSliderChange() {
    trackMixpanelEvent('scope.time.slider.change', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  render() {
    // Don't render the time travel control if it's not explicitly enabled for this instance.
    if (!this.props.showingTimeTravel) return null;

    const { sliderValue, sliderMinValue, inputValue } = this.state;
    const sliderMaxValue = moment().valueOf();

    return (
      <div className="time-travel">
        <div className="time-travel-slider-wrapper">
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
          <a className="button jump" onClick={() => this.jumpInTime(FIVE_MINUTES_MS)}>
            <span className="fa fa-step-forward" /> 5 mins
          </a>
          <a className="button jump" onClick={() => this.jumpInTime(ONE_HOUR_MS)}>
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
