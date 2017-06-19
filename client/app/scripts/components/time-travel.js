import React from 'react';
import moment from 'moment';
import Slider from 'rc-slider';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import TimeTravelTimestamp from './time-travel-timestamp';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import {
  websocketQueryInPast,
  startWebsocketTransitionLoader,
  clickResumeUpdate,
} from '../actions/app-actions';

import {
  TIMELINE_SLIDER_UPDATE_INTERVAL,
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';


const sliderRanges = {
  last15Minutes: {
    label: 'Last 15 minutes',
    getStart: () => moment().utc().subtract(15, 'minutes'),
  },
  last1Hour: {
    label: 'Last 1 hour',
    getStart: () => moment().utc().subtract(1, 'hour'),
  },
  last6Hours: {
    label: 'Last 6 hours',
    getStart: () => moment().utc().subtract(6, 'hours'),
  },
  last24Hours: {
    label: 'Last 24 hours',
    getStart: () => moment().utc().subtract(24, 'hours'),
  },
  last7Days: {
    label: 'Last 7 days',
    getStart: () => moment().utc().subtract(7, 'days'),
  },
  last30Days: {
    label: 'Last 30 days',
    getStart: () => moment().utc().subtract(30, 'days'),
  },
  last90Days: {
    label: 'Last 90 days',
    getStart: () => moment().utc().subtract(90, 'days'),
  },
  last1Year: {
    label: 'Last 1 year',
    getStart: () => moment().subtract(1, 'year'),
  },
  todaySoFar: {
    label: 'Today so far',
    getStart: () => moment().utc().startOf('day'),
  },
  thisWeekSoFar: {
    label: 'This week so far',
    getStart: () => moment().utc().startOf('week'),
  },
  thisMonthSoFar: {
    label: 'This month so far',
    getStart: () => moment().utc().startOf('month'),
  },
  thisYearSoFar: {
    label: 'This year so far',
    getStart: () => moment().utc().startOf('year'),
  },
};

class TimeTravel extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      showSliderPanel: false,
      millisecondsInPast: 0,
      rangeOptionSelected: sliderRanges.last1Hour,
    };

    this.renderRangeOption = this.renderRangeOption.bind(this);
    this.handleTimestampClick = this.handleTimestampClick.bind(this);
    this.handleJumpToNowClick = this.handleJumpToNowClick.bind(this);
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
    let millisecondsInPast = this.getRangeMilliseconds() - sliderValue;

    // If the slider value is less than 1s away from the right-end (current time),
    // assume we meant the current time - this is important for the '... so far'
    // ranges where the range of values changes over time.
    if (millisecondsInPast < 1000) {
      millisecondsInPast = 0;
    }

    this.setState({ millisecondsInPast });
    this.props.startWebsocketTransitionLoader();
    this.debouncedUpdateTimestamp(millisecondsInPast);

    this.debouncedTrackSliderChange();
  }

  handleRangeOptionClick(rangeOption) {
    this.setState({ rangeOptionSelected: rangeOption });

    const rangeMilliseconds = this.getRangeMilliseconds(rangeOption);
    if (this.state.millisecondsInPast > rangeMilliseconds) {
      this.setState({ millisecondsInPast: rangeMilliseconds });
      this.updateTimestamp(rangeMilliseconds);
      this.props.startWebsocketTransitionLoader();
    }

    trackMixpanelEvent('scope.time.range.select', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      label: rangeOption.label,
    });
  }

  handleJumpToNowClick() {
    this.setState({
      showSliderPanel: false,
      millisecondsInPast: 0,
      rangeOptionSelected: sliderRanges.last1Hour,
    });
    this.updateTimestamp();
    this.props.startWebsocketTransitionLoader();

    trackMixpanelEvent('scope.time.now.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  handleTimestampClick() {
    const showSliderPanel = !this.state.showSliderPanel;
    this.setState({ showSliderPanel });

    trackMixpanelEvent('scope.time.timestamp.click', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      showSliderPanel,
    });
  }

  updateTimestamp(millisecondsInPast = 0) {
    this.props.websocketQueryInPast(millisecondsInPast);
    this.props.clickResumeUpdate();
  }

  getRangeMilliseconds(rangeOption = this.state.rangeOptionSelected) {
    return moment().diff(rangeOption.getStart());
  }

  trackSliderChange() {
    trackMixpanelEvent('scope.time.slider.change', {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  renderRangeOption(rangeOption) {
    const handleClick = () => { this.handleRangeOptionClick(rangeOption); };
    const selected = (this.state.rangeOptionSelected.label === rangeOption.label);
    const className = classNames('option', { selected });

    return (
      <a key={rangeOption.label} className={className} onClick={handleClick}>
        {rangeOption.label}
      </a>
    );
  }

  renderJumpToNowButton() {
    return (
      <a className="button jump-to-now" title="Jump to now" onClick={this.handleJumpToNowClick}>
        <span className="fa fa-step-forward" />
      </a>
    );
  }

  renderTimeSlider() {
    const { millisecondsInPast } = this.state;
    const rangeMilliseconds = this.getRangeMilliseconds();

    return (
      <Slider
        onChange={this.handleSliderChange}
        value={rangeMilliseconds - millisecondsInPast}
        max={rangeMilliseconds}
      />
    );
  }

  render() {
    const { websocketTransitioning, hasTimeTravel } = this.props;
    const { showSliderPanel, millisecondsInPast, rangeOptionSelected } = this.state;
    const lowerCaseLabel = rangeOptionSelected.label.toLowerCase();
    const isCurrent = (millisecondsInPast === 0);

    // Don't render the time travel control if it's not explicitly enabled for this instance.
    if (!hasTimeTravel) return null;

    return (
      <div className="time-travel">
        {showSliderPanel && <div className="time-travel-slider">
          <div className="options">
            <div className="column">
              {this.renderRangeOption(sliderRanges.last15Minutes)}
              {this.renderRangeOption(sliderRanges.last1Hour)}
              {this.renderRangeOption(sliderRanges.last6Hours)}
              {this.renderRangeOption(sliderRanges.last24Hours)}
            </div>
            <div className="column">
              {this.renderRangeOption(sliderRanges.last7Days)}
              {this.renderRangeOption(sliderRanges.last30Days)}
              {this.renderRangeOption(sliderRanges.last90Days)}
              {this.renderRangeOption(sliderRanges.last1Year)}
            </div>
            <div className="column">
              {this.renderRangeOption(sliderRanges.todaySoFar)}
              {this.renderRangeOption(sliderRanges.thisWeekSoFar)}
              {this.renderRangeOption(sliderRanges.thisMonthSoFar)}
              {this.renderRangeOption(sliderRanges.thisYearSoFar)}
            </div>
          </div>
          <span className="slider-tip">Move the slider to explore {lowerCaseLabel}</span>
          {this.renderTimeSlider()}
        </div>}
        <div className="time-travel-status">
          {websocketTransitioning && <div className="time-travel-jump-loader">
            <span className="fa fa-circle-o-notch fa-spin" />
          </div>}
          <TimeTravelTimestamp
            onClick={this.handleTimestampClick}
            millisecondsInPast={millisecondsInPast}
            selected={showSliderPanel}
          />
          {!isCurrent && this.renderJumpToNowButton()}
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
    websocketTransitioning: scope.get('websocketTransitioning'),
    topologyViewMode: scope.get('topologyViewMode'),
    currentTopology: scope.get('currentTopology'),
  };
}

export default connect(
  mapStateToProps,
  {
    websocketQueryInPast,
    startWebsocketTransitionLoader,
    clickResumeUpdate,
  }
)(TimeTravel);
