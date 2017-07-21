import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { debounce } from 'lodash';
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

const yFunc = (scale, start) => {
  const end = start * 1.4;
  if (scale < start) return 0;
  if (scale > end) return 1;
  return (Math.log(scale) - Math.log(start)) / (Math.log(end) - Math.log(start));
};

const getShift = (period, scale) => {
  const yearShift = 1;
  const monthShift = yFunc(scale, 0.0052);
  const dayShift = yFunc(scale, 0.067);
  const minuteShift = yFunc(scale, 1.9);
  // console.log(scale, yearShift, monthShift, dayShift, minuteShift);
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

function scaleDuration(duration, scale) {
  return moment.duration(duration.asMilliseconds() * scale);
}

const WIDTH = 5000;
const C = 2000000;
const MIN_TICK_SPACING = 40;
const MIN_ZOOM = 0.002;
const MAX_ZOOM = 1000;

class TimeTravelTimeline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      timestampNow: moment(),
      focusedTimestamp: moment(),
      timelineRange: moment.duration(C, 'seconds'),
      isDragging: false,
      scaleX: 1,
    };

    this.width = 2000;

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
  }

  updateTimestamp(timestamp) {
    this.props.jumpToTime(moment(timestamp));
  }

  handleZoom(e) {
    let scaleX = this.state.scaleX * Math.pow(1.002, -e.deltaY);
    scaleX = Math.min(Math.max(scaleX, MIN_ZOOM), MAX_ZOOM);
    const timelineRange = moment.duration(C / scaleX, 'seconds');
    this.setState({ timelineRange, scaleX });
  }

  handlePanning() {
    // let scaleX = this.state.scaleX * Math.pow(1.00, (d3Event.dx === 0 ? -d3Event.dy : 0));
    // scaleX = Math.max(Math.min(scaleX, MAX_ZOOM), MIN_ZOOM);
    // const timelineRange = moment.duration(C / scaleX, 'seconds');
    // this.setState({ timelineRange, scaleX });

    const { focusedTimestamp, timelineRange } = this.state;
    const dragDuration = scaleDuration(timelineRange, -d3Event.dx / WIDTH);
    const newTimestamp = moment(focusedTimestamp).add(dragDuration);
    this.jumpTo(newTimestamp);
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.setState({ isPanning: false });
  }

  jumpTo(timestamp) {
    const { timestampNow } = this.state;
    const focusedTimestamp = timestamp > timestampNow ? timestampNow : timestamp;
    this.props.onUpdateTimestamp(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  jumpForward() {
    const { focusedTimestamp, timelineRange } = this.state;
    const duration = scaleDuration(timelineRange, this.width / WIDTH / 2);
    const timestamp = moment(focusedTimestamp).add(duration);
    this.jumpTo(timestamp);
  }

  jumpBackward() {
    const { focusedTimestamp, timelineRange } = this.state;
    const duration = scaleDuration(timelineRange, this.width / WIDTH / 2);
    const timestamp = moment(focusedTimestamp).subtract(duration);
    this.jumpTo(timestamp);
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  getTimeScale() {
    const { timelineRange, focusedTimestamp } = this.state;
    const timelineHalfRange = scaleDuration(timelineRange, 0.5);
    const roundedTimestamp = moment(focusedTimestamp).startOf('second').utc();
    const startDate = moment(roundedTimestamp).subtract(timelineHalfRange);
    const endDate = moment(roundedTimestamp).add(timelineHalfRange);
    return scaleUtc()
      .domain([startDate, endDate])
      .range([-WIDTH / 2, WIDTH / 2]);
  }

  timestampTransform(timestamp) {
    const timeScale = this.getTimeScale();
    return `translate(${timeScale(timestamp)}, 0)`;
  }

  renderPeriodBar(period, prevPeriod, periodFormat, durations) {
    const { timelineRange } = this.state;
    const timeScale = this.getTimeScale();

    let duration = null;
    for (let i = 0; i < durations.length; i += 1) {
      if (timelineRange < scaleDuration(durations[i], 65)) {
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
        while (ts.length > 0 && p - timeScale(ts[ts.length - 1]) < MIN_TICK_SPACING) {
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

    const p = getShift(period, this.state.scaleX);
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
    const { timestampNow } = this.state;
    return (
      <rect
        className="available-range"
        transform={this.timestampTransform(timestampNow)}
        x={-this.width} width={this.width} height={70}
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
