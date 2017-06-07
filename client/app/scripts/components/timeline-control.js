import React from 'react';
import moment from 'moment';
import Slider from 'rc-slider';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import PauseButton from './pause-button';
import TopologyTimestampButton from './topology-timestamp-button';
import {
  websocketQueryTimestamp,
  clickResumeUpdate,
  startMovingInTime,
} from '../actions/app-actions';

import { TIMELINE_DEBOUNCE_INTERVAL } from '../constants/timer';


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

class TimelineControl extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      showTimelinePanel: false,
      millisecondsInPast: 0,
      rangeOptionSelected: sliderRanges.last1Hour,
    };

    this.jumpToNow = this.jumpToNow.bind(this);
    this.toggleTimelinePanel = this.toggleTimelinePanel.bind(this);
    this.handleSliderChange = this.handleSliderChange.bind(this);
    this.renderRangeOption = this.renderRangeOption.bind(this);
    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentWillUnmount() {
    this.updateTimestamp(null);
  }

  updateTimestamp(timestampSinceNow) {
    this.props.websocketQueryTimestamp(timestampSinceNow);
    this.props.clickResumeUpdate();
  }

  toggleTimelinePanel() {
    this.setState({ showTimelinePanel: !this.state.showTimelinePanel });
  }

  handleSliderChange(sliderValue) {
    const millisecondsInPast = this.getRangeMilliseconds() - sliderValue;
    this.props.startMovingInTime();
    this.debouncedUpdateTimestamp(millisecondsInPast);
    this.setState({ millisecondsInPast });
  }

  handleRangeOptionClick(rangeOption) {
    this.setState({ rangeOptionSelected: rangeOption });

    const rangeMilliseconds = this.getRangeMilliseconds(rangeOption);
    if (this.state.millisecondsInPast > rangeMilliseconds) {
      this.updateTimestamp(rangeMilliseconds);
      this.setState({ millisecondsInPast: rangeMilliseconds });
    }
  }

  getRangeMilliseconds(rangeOption) {
    rangeOption = rangeOption || this.state.rangeOptionSelected;
    return moment().diff(rangeOption.getStart());
  }

  jumpToNow() {
    this.setState({
      showTimelinePanel: false,
      millisecondsInPast: 0,
      rangeOptionSelected: sliderRanges.last1Hour,
    });
    this.props.startMovingInTime();
    this.updateTimestamp(null);
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
      <a className="button jump-to-now" title="Jump to now" onClick={this.jumpToNow}>
        <span className="fa fa-step-forward" />
      </a>
    );
  }

  renderTimelineSlider() {
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
    const { movingInTime } = this.props;
    const { showTimelinePanel, millisecondsInPast } = this.state;
    const isCurrent = (millisecondsInPast === 0);

    return (
      <div className="timeline-control">
        {showTimelinePanel && <div className="timeline-panel">
          <span className="caption">Explore</span>
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
          <span className="slider-tip">Move the slider to travel back in time</span>
          {this.renderTimelineSlider()}
        </div>}
        <div className="time-status">
          {movingInTime && <div className="timeline-jump-loader">
            <span className="fa fa-circle-o-notch fa-spin" />
          </div>}
          <TopologyTimestampButton
            onClick={this.toggleTimelinePanel}
            millisecondsInPast={millisecondsInPast}
            selected={showTimelinePanel}
          />
          {!isCurrent && this.renderJumpToNowButton()}
          <PauseButton />
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    movingInTime: state.get('websocketMovingInTime'),
  };
}

export default connect(
  mapStateToProps,
  {
    websocketQueryTimestamp,
    clickResumeUpdate,
    startMovingInTime,
  }
)(TimelineControl);
