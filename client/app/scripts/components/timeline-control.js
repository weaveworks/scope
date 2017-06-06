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

class TimelineControl extends React.Component {
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
    this.updateTimestamp(null);
  }

  updateTimestamp(timestampSinceNow) {
    this.props.websocketQueryTimestamp(timestampSinceNow);
    this.props.clickResumeUpdate();
  }

  toggleTimelinePanel() {
    this.setState({ showTimelinePanel: !this.state.showTimelinePanel });
  }

  handleSliderChange(value) {
    const offsetMilliseconds = this.getRangeMilliseconds() - value;
    this.props.startMovingInTime();
    this.debouncedUpdateTimestamp(offsetMilliseconds);
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
    this.updateTimestamp(null);
  }

  getTotalOffset() {
    const { rangeOptionSelected, offsetMilliseconds } = this.state;
    const rangeBehindMilliseconds = moment().utc().diff(rangeOptionSelected.getEnd());
    return offsetMilliseconds + rangeBehindMilliseconds;
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

  renderJumpToNowButton() {
    return (
      <a className="button jump-to-now" title="Jump to now" onClick={this.jumpToNow}>
        <span className="fa fa-step-forward" />
      </a>
    );
  }

  renderTimelineSlider() {
    const { offsetMilliseconds } = this.state;
    const rangeMilliseconds = this.getRangeMilliseconds();

    return (
      <Slider
        onChange={this.handleSliderChange}
        value={rangeMilliseconds - offsetMilliseconds}
        max={rangeMilliseconds}
      />
    );
  }

  render() {
    const { movingInTime } = this.props;
    const { showTimelinePanel, offsetMilliseconds } = this.state;

    const showingCurrent = (this.getTotalOffset() === 0);

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
            selected={showTimelinePanel}
            offset={offsetMilliseconds}
          />
          {!showingCurrent && this.renderJumpToNowButton()}
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
