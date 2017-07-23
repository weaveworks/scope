import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { debounce, clamp } from 'lodash';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import { drag } from 'd3-drag';
import { scaleUtc } from 'd3-scale';
import { event as d3Event, select } from 'd3-selection';
import {
  jumpToTime,
} from '../actions/app-actions';

import {
  TIMELINE_TICK_INTERVAL,
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';

const MINUTE_INTERVALS = [
  moment.duration(1, 'minute'),
  moment.duration(5, 'minutes'),
  moment.duration(15, 'minutes'),
  moment.duration(1, 'hour'),
  moment.duration(3, 'hours'),
  moment.duration(6, 'hours'),
];
const DAY_INTERVALS = [
  moment.duration(1, 'day'),
  moment.duration(1, 'week'),
];
const MONTH_INTERVALS = [
  moment.duration(1, 'month'),
  moment.duration(3, 'months'),
];
const YEAR_INTERVALS = [
  moment.duration(1, 'year'),
];

function scaleDuration(duration, scale) {
  return moment.duration(duration.asMilliseconds() * scale);
}

function isShorterThan(duration1, duration2) {
  return duration1.asMilliseconds() < duration2.asMilliseconds();
}

function isLongerThan(duration1, duration2) {
  return duration1.asMilliseconds() > duration2.asMilliseconds();
}

const WIDTH = 5000;
const FADE_OUT_FACTOR = 1.4;
const MIN_TICK_SPACING = 70;
const MAX_TICK_SPACING = 415;
const MIN_DURATION = moment.duration(250, 'milliseconds');
const MAX_DURATION = moment.duration(3, 'days');

const yFunc = (currentDuration, fadedInDuration) => {
  const durationLog = d => Math.log(d.asMilliseconds());
  const fadedOutDuration = scaleDuration(fadedInDuration, FADE_OUT_FACTOR);
  const transitionFactor = durationLog(fadedOutDuration) - durationLog(currentDuration);
  const transitionLength = durationLog(fadedOutDuration) - durationLog(fadedInDuration);
  return clamp(transitionFactor / transitionLength, 0, 1);
};

const getShift = (period, duration) => {
  const durationMultiplier = 1 / MAX_TICK_SPACING;
  const minuteShift = yFunc(duration, scaleDuration(moment.duration(1, 'day'), durationMultiplier));
  const dayShift = yFunc(duration, scaleDuration(moment.duration(1, 'month'), durationMultiplier));
  const monthShift = yFunc(duration, scaleDuration(moment.duration(1, 'year'), durationMultiplier));
  const yearShift = 1;

  let result = 0;
  switch (period) {
    case 'year': result = yearShift + monthShift + dayShift + minuteShift; break;
    case 'month': result = monthShift + dayShift + minuteShift; break;
    case 'day': result = dayShift + minuteShift; break;
    case 'minute': result = minuteShift; break;
    default: result = 0; break;
  }
  return result;
};

class TimeTravelTimeline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      timestampNow: moment(),
      focusedTimestamp: moment(),
      durationPerPixel: moment.duration(1, 'minute'),
      isDragging: false,
    };

    this.width = 0;
    this.height = 0;

    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.jumpTo = this.jumpTo.bind(this);
    this.jumpForward = this.jumpForward.bind(this);
    this.jumpBackward = this.jumpBackward.bind(this);

    this.handlePanning = this.handlePanning.bind(this);
    this.handlePanStart = this.handlePanStart.bind(this);
    this.handlePanEnd = this.handlePanEnd.bind(this);
    this.handleZoom = this.handleZoom.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    this.svg = select('svg#time-travel-timeline');
    this.drag = drag()
      .on('start', this.handlePanStart)
      .on('end', this.handlePanEnd)
      .on('drag', this.handlePanning);
    this.svg.call(this.drag);

    // Force periodic re-renders to update the slider position as time goes by.
    this.timer = setInterval(() => {
      this.setState({ timestampNow: moment().startOf('second') });
    }, TIMELINE_TICK_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.pausedAt) {
      this.setState({ focusedTimestamp: nextProps.pausedAt });
    }
    this.width = this.svgRef.getBoundingClientRect().width;
    this.height = this.svgRef.getBoundingClientRect().height;
  }

  updateTimestamp(timestamp) {
    this.props.jumpToTime(moment(timestamp));
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.setState({ isPanning: false });
  }

  handlePanning() {
    const { focusedTimestamp, durationPerPixel } = this.state;
    const dragDuration = scaleDuration(durationPerPixel, d3Event.dx);
    const newTimestamp = moment(focusedTimestamp).subtract(dragDuration);
    this.jumpTo(newTimestamp);
  }

  handleZoom(e) {
    const scale = Math.pow(1.0015, e.deltaY);
    let durationPerPixel = scaleDuration(this.state.durationPerPixel, scale);
    if (isLongerThan(durationPerPixel, MAX_DURATION)) durationPerPixel = MAX_DURATION;
    if (isShorterThan(durationPerPixel, MIN_DURATION)) durationPerPixel = MIN_DURATION;
    this.setState({ durationPerPixel });
  }

  jumpTo(timestamp) {
    const { timestampNow } = this.state;
    const focusedTimestamp = timestamp > timestampNow ? timestampNow : timestamp;
    this.props.onUpdateTimestamp(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  jumpForward() {
    const { focusedTimestamp, durationPerPixel } = this.state;
    const duration = scaleDuration(durationPerPixel, this.width / 4);
    const timestamp = moment(focusedTimestamp).add(duration);
    this.jumpTo(timestamp);
  }

  jumpBackward() {
    const { focusedTimestamp, durationPerPixel } = this.state;
    const duration = scaleDuration(durationPerPixel, this.width / 4);
    const timestamp = moment(focusedTimestamp).subtract(duration);
    this.jumpTo(timestamp);
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  getTimeScale() {
    const { durationPerPixel, focusedTimestamp } = this.state;
    const timelineHalfRange = scaleDuration(durationPerPixel, 0.5);
    const roundedTimestamp = moment(focusedTimestamp).startOf('second').utc();
    const startDate = moment(roundedTimestamp).subtract(timelineHalfRange);
    const endDate = moment(roundedTimestamp).add(timelineHalfRange);
    return scaleUtc()
      .domain([startDate, endDate])
      .range([-0.5, 0.5]);
  }

  timestampTransform(timestamp) {
    const timeScale = this.getTimeScale();
    return `translate(${timeScale(timestamp)}, 0)`;
  }

  renderPeriodBar(period, prevPeriod, periodFormat, durations) {
    const { durationPerPixel } = this.state;
    const timeScale = this.getTimeScale();

    let duration = null;
    for (let i = 0; i < durations.length; i += 1) {
      if (scaleDuration(durationPerPixel, MIN_TICK_SPACING) < durations[i]) {
        duration = durations[i];
        break;
      }
    }

    if (!duration) return null;

    const startDate = moment(timeScale.invert(-WIDTH / 2));
    let t = moment(startDate).startOf(prevPeriod);
    let turningPoint = moment(t).add(1, prevPeriod);
    const ts = [];

    do {
      const p = timeScale(t);
      if (p > -this.width && p < this.width) {
        while (ts.length > 0 && p - timeScale(ts[ts.length - 1]) < 0.85 * MIN_TICK_SPACING) {
          ts.pop();
        }
        ts.push(t);
      }
      t = moment(t).add(duration);
      if (prevPeriod !== period && t >= turningPoint) {
        t = turningPoint;
        turningPoint = moment(turningPoint).add(1, prevPeriod);
      }
    } while (timeScale(t) < this.width);

    const p = getShift(period, this.state.durationPerPixel);
    const shift = 2 + (55 * (1 - (p * 0.25)));
    const opacity = Math.min(p * p, 1);
    return (
      <g className={period} transform={`translate(0, ${shift})`} style={{ opacity }}>
        {fromJS(ts).map(timestamp => (
          <g transform={this.timestampTransform(timestamp)} key={timestamp.format()}>
            <line y2="75" stroke="#ddd" strokeWidth="1" />
            <foreignObject width="100" height="20">
              <a className="timestamp-label" onClick={() => this.jumpTo(timestamp)}>
                {timestamp.format(periodFormat)}
              </a>
            </foreignObject>
          </g>
        ))}
      </g>
    );
  }

  renderAvailableRange() {
    const timeScale = this.getTimeScale();
    const nowShift = Math.min(timeScale(this.state.timestampNow), WIDTH / 2);
    return (
      <rect
        className="available-range"
        transform={`translate(${nowShift}, 0)`}
        x={-WIDTH} width={WIDTH} height={this.height}
      />
    );
  }

  renderAxis() {
    return (
      <g id="axis">
        {this.renderAvailableRange()}
        <g className="ticks">
          {this.renderPeriodBar('year', 'year', 'YYYY', YEAR_INTERVALS)}
          {this.renderPeriodBar('month', 'year', 'MMMM', MONTH_INTERVALS)}
          {this.renderPeriodBar('day', 'month', 'Do', DAY_INTERVALS)}
          {this.renderPeriodBar('minute', 'day', 'HH:mm', MINUTE_INTERVALS)}
        </g>
      </g>
    );
  }

  render() {
    const className = classNames({ dragging: this.state.isPanning });
    return (
      <div className="time-travel-timeline">
        <a className="button jump-backward" onClick={this.jumpBackward}>
          <span className="fa fa-chevron-left" />
        </a>
        <svg id="time-travel-timeline" className={className} ref={this.saveSvgRef} onWheel={this.handleZoom}>
          <g className="view" transform={`translate(${this.width / 2}, 0)`}>
            {this.renderAxis()}
          </g>
        </svg>
        <a className="button jump-forward" onClick={this.jumpForward}>
          <span className="fa fa-chevron-right" />
        </a>
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    viewportWidth: state.getIn(['viewport', 'width']),
    pausedAt: state.get('pausedAt'),
  };
}


export default connect(
  mapStateToProps,
  {
    jumpToTime,
  }
)(TimeTravelTimeline);
