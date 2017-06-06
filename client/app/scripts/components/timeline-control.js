import React from 'react';
import moment from 'moment';
import Slider from 'rc-slider';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import PauseButton from './pause-button';
import TopologyTimestampInfo from './topology-timestamp-info';
import { websocketQueryTimestamp, startMovingInTime } from '../actions/app-actions';
import { TIMELINE_DEBOUNCE_INTERVAL } from '../constants/timer';


const sliderRanges = {
  last15Minutes: {
    label: 'Last 15 minutes',
    getStart: () => moment().utc().subtract(15, 'minutes'),
    getEnd: () => moment().utc(),
  },
  last1Hour: {
    label: 'Last 1 hour',
    getStart: () => moment().utc().subtract(1, 'hour'),
    getEnd: () => moment().utc(),
  },
  last6Hours: {
    label: 'Last 6 hours',
    getStart: () => moment().utc().subtract(6, 'hours'),
    getEnd: () => moment().utc(),
  },
  last24Hours: {
    label: 'Last 24 hours',
    getStart: () => moment().utc().subtract(24, 'hours'),
    getEnd: () => moment().utc(),
  },
  last7Days: {
    label: 'Last 7 days',
    getStart: () => moment().utc().subtract(7, 'days'),
    getEnd: () => moment().utc(),
  },
  last30Days: {
    label: 'Last 30 days',
    getStart: () => moment().utc().subtract(30, 'days'),
    getEnd: () => moment().utc(),
  },
  last90Days: {
    label: 'Last 90 days',
    getStart: () => moment().utc().subtract(90, 'days'),
    getEnd: () => moment().utc(),
  },
  last1Year: {
    label: 'Last 1 year',
    getStart: () => moment().subtract(1, 'year'),
    getEnd: () => moment().utc(),
  },
  todaySoFar: {
    label: 'Today so far',
    getStart: () => moment().utc().startOf('day'),
    getEnd: () => moment().utc(),
  },
  thisWeekSoFar: {
    label: 'This week so far',
    getStart: () => moment().utc().startOf('week'),
    getEnd: () => moment().utc(),
  },
  thisMonthSoFar: {
    label: 'This month so far',
    getStart: () => moment().utc().startOf('month'),
    getEnd: () => moment().utc(),
  },
  thisYearSoFar: {
    label: 'This year so far',
    getStart: () => moment().utc().startOf('year'),
    getEnd: () => moment().utc(),
  },
  // yesterday: {
  //   label: 'Yesterday',
  //   getStart: () => moment().utc().subtract(1, 'day').startOf('day'),
  //   getEnd: () => moment().utc().subtract(1, 'day').endOf('day'),
  // },
  // previousWeek: {
  //   label: 'Previous week',
  //   getStart: () => moment().utc().subtract(1, 'week').startOf('week'),
  //   getEnd: () => moment().utc().subtract(1, 'week').endOf('week'),
  // },
  // previousMonth: {
  //   label: 'Previous month',
  //   getStart: () => moment().utc().subtract(1, 'month').startOf('month'),
  //   getEnd: () => moment().utc().subtract(1, 'month').endOf('month'),
  // },
  // previousYear: {
  //   label: 'Previous year',
  //   getStart: () => moment().utc().subtract(1, 'year').startOf('year'),
  //   getEnd: () => moment().utc().subtract(1, 'year').endOf('year'),
  // },
};

class TimelineControl extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = {
      showTimelinePanel: false,
      offsetMilliseconds: 0,
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
    this.updateTimestamp(moment());
  }

  updateTimestamp(timestamp) {
    this.props.websocketQueryTimestamp(timestamp);
  }

  toggleTimelinePanel() {
    this.setState({ showTimelinePanel: !this.state.showTimelinePanel });
  }

  handleSliderChange(value) {
    const offsetMilliseconds = this.getRangeMilliseconds() - value;
    const timestamp = moment().utc().subtract(offsetMilliseconds);
    this.props.startMovingInTime();
    this.debouncedUpdateTimestamp(timestamp);
    this.setState({ offsetMilliseconds });
  }

  getRangeMilliseconds() {
    const range = this.state.rangeOptionSelected;
    return range.getEnd().diff(range.getStart());
  }

  jumpToNow() {
    this.setState({
      showTimelinePanel: false,
      offsetMilliseconds: 0,
      rangeOptionSelected: sliderRanges.last1Hour,
    });
    this.props.startMovingInTime();
    this.updateTimestamp(moment());
  }

  renderRangeOption(option) {
    const handleClick = () => { this.setState({ rangeOptionSelected: option }); };
    const selected = (this.state.rangeOptionSelected.label === option.label);
    const className = classNames('option', { selected });

    return (
      <a key={option.label} className={className} onClick={handleClick}>
        {option.label}
      </a>
    );
  }

  getTotalOffset() {
    const { rangeOptionSelected, offsetMilliseconds } = this.state;
    const rangeBehindMilliseconds = moment().diff(rangeOptionSelected.getEnd());
    return offsetMilliseconds + rangeBehindMilliseconds;
  }

  render() {
    const { showTimelinePanel, offsetMilliseconds } = this.state;
    const rangeMilliseconds = this.getRangeMilliseconds();

    const showingCurrent = (this.getTotalOffset() === 0);
    const timeStatusClassName = classNames('time-status', { 'showing-current': showingCurrent });
    const toggleButtonClassName = classNames('button toggle', { selected: showTimelinePanel });

    return (
      <div className="timeline-control">
        {showTimelinePanel && <div className="timeline-panel">
          <strong>Explore</strong>
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
          <strong>Move the slider to travel in time</strong>
          <Slider
            onChange={this.handleSliderChange}
            value={rangeMilliseconds - offsetMilliseconds}
            max={rangeMilliseconds}
          />
        </div>}
        <div className={timeStatusClassName}>
          <a className={toggleButtonClassName} onClick={this.toggleTimelinePanel}>
            <TopologyTimestampInfo />
            <span className="fa fa-clock-o" />
          </a>
          <PauseButton />
          {!showingCurrent && <a
            className="button jump-to-now"
            title="Jump to now"
            onClick={this.jumpToNow}>
            <span className="fa fa-step-forward" />
          </a>}
        </div>
      </div>
    );
  }
}

export default connect(
  null,
  {
    websocketQueryTimestamp,
    startMovingInTime,
  }
)(TimelineControl);
