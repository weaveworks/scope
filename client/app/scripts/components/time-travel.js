import React from 'react';
import Slider from 'rc-slider';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import TimeTravelTimestamp from './time-travel-timestamp';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import {
  timeTravelJumpToPast,
  timeTravelStartTransition,
} from '../actions/app-actions';

import {
  TIMELINE_SLIDER_UPDATE_INTERVAL,
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';


function getRangeMilliseconds() {
  return 90 * 24 * 60 * 60 * 1000;
}

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      showSliderPanel: false,
      millisecondsInPast: 0,
    };

    this.handleTimestampClick = this.handleTimestampClick.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
    this.debouncedTrackSliderChange = debounce(
      this.trackSliderChange.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    // Force periodic re-renders to update the slider position as time goes by.
    this.timer = setInterval(() => { this.forceUpdate(); }, TIMELINE_SLIDER_UPDATE_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
    this.updateTimestamp();
  }

  handleSliderChange(sliderValue) {
    let millisecondsInPast = getRangeMilliseconds() - sliderValue;

    // If the slider value is less than 1s away from the right-end (current time),
    // assume we meant the current time - this is important for the '... so far'
    // ranges where the range of values changes over time.
    if (millisecondsInPast < 1000) {
      millisecondsInPast = 0;
    }

    this.setState({ millisecondsInPast });
    this.props.timeTravelStartTransition();
    this.debouncedUpdateTimestamp(millisecondsInPast);

    this.debouncedTrackSliderChange();
  }

  handleTimestampClick() {
    trackMixpanelEvent('scope.time.timestamp.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  updateTimestamp(millisecondsInPast = 0) {
    this.props.timeTravelJumpToPast(millisecondsInPast);
  }

  trackSliderChange() {
    trackMixpanelEvent('scope.time.slider.change', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  render() {
    const { timeTravelTransitioning, hasTimeTravel } = this.props;
    const { showSliderPanel, millisecondsInPast } = this.state;
    const rangeMilliseconds = getRangeMilliseconds();

    // Don't render the time travel control if it's not explicitly enabled for this instance.
    if (!hasTimeTravel) return null;

    return (
      <div className="time-travel">
        <div className="time-travel-slider-wrapper">
          <Slider
            onChange={this.handleSliderChange}
            value={rangeMilliseconds - millisecondsInPast}
            max={rangeMilliseconds}
          />
        </div>
        <div className="time-travel-timestamp-controls-wrapper">
          {timeTravelTransitioning && <div className="time-travel-jump-loader">
            <span className="fa fa-circle-o-notch fa-spin" />
          </div>}
          <TimeTravelTimestamp
            onClick={this.handleTimestampClick}
            millisecondsInPast={millisecondsInPast}
            selected={showSliderPanel}
          />
        </div>
      </div>
    );
  }
}

function mapStateToProps({ scope, root }, { params }) {
  const cloudInstance = root.instances[params.orgId] || {};
  const featureFlags = cloudInstance.featureFlags || [];
  return {
    hasTimeTravel: featureFlags.includes('time-travel'),
    timeTravelTransitioning: scope.get('timeTravelTransitioning'),
    topologyViewMode: scope.get('topologyViewMode'),
    currentTopology: scope.get('currentTopology'),
  };
}

export default connect(
  mapStateToProps,
  {
    timeTravelJumpToPast,
    timeTravelStartTransition,
  }
)(TimeTravel);
