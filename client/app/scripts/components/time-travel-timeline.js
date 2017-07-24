import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { debounce, map, clamp, find, last } from 'lodash';
import { connect } from 'react-redux';
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


const TICK_ROWS = {
  minute: {
    format: 'HH:mm',
    intervals: [
      moment.duration(1, 'minute'),
      moment.duration(5, 'minutes'),
      moment.duration(15, 'minutes'),
      moment.duration(1, 'hour'),
      moment.duration(3, 'hours'),
      moment.duration(6, 'hours'),
    ],
  },
  day: {
    format: 'Do',
    intervals: [
      moment.duration(1, 'day'),
      moment.duration(1, 'week'),
    ],
  },
  month: {
    format: 'MMMM',
    intervals: [
      moment.duration(1, 'month'),
      moment.duration(3, 'months'),
    ],
  },
  year: {
    format: 'YYYY',
    intervals: [
      moment.duration(1, 'year'),
    ],
  },
};

const FADE_OUT_FACTOR = 1.4;
const MIN_TICK_SPACING_PX = 70;
const MAX_TICK_SPACING_PX = 415;
const MIN_DURATION_PER_PX = moment.duration(250, 'milliseconds');
const INIT_DURATION_PER_PX = moment.duration(1, 'minute');
const MAX_DURATION_PER_PX = moment.duration(3, 'days');

function scaleDuration(duration, scale) {
  return moment.duration(duration.asMilliseconds() * scale);
}

function yFunc(currentDuration, fadedInDuration) {
  const durationLog = d => Math.log(d.asMilliseconds());
  const fadedOutDuration = scaleDuration(fadedInDuration, FADE_OUT_FACTOR);
  const transitionFactor = durationLog(fadedOutDuration) - durationLog(currentDuration);
  const transitionLength = durationLog(fadedOutDuration) - durationLog(fadedInDuration);
  return clamp(transitionFactor / transitionLength, 0, 1);
}

function getFadeInFactor(period, duration) {
  const durationMultiplier = 1 / MAX_TICK_SPACING_PX;
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
}


class TimeTravelTimeline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      timestampNow: moment(),
      focusedTimestamp: moment(),
      durationPerPixel: INIT_DURATION_PER_PX,
      boundingRect: { width: 0, height: 0 },
      isDragging: false,
    };

    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.jumpTo = this.jumpTo.bind(this);
    this.jumpForward = this.jumpForward.bind(this);
    this.jumpBackward = this.jumpBackward.bind(this);

    this.findOptimalDuration = this.findOptimalDuration.bind(this);

    this.handlePanStart = this.handlePanStart.bind(this);
    this.handlePanEnd = this.handlePanEnd.bind(this);
    this.handlePan = this.handlePan.bind(this);
    this.handleZoom = this.handleZoom.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    this.svg = select('.time-travel-timeline svg');
    this.drag = drag()
      .on('start', this.handlePanStart)
      .on('end', this.handlePanEnd)
      .on('drag', this.handlePan);
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
    this.setState({ boundingRect: this.svgRef.getBoundingClientRect() });
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

  handlePan() {
    const { focusedTimestamp, durationPerPixel } = this.state;
    const dragDuration = scaleDuration(durationPerPixel, d3Event.dx);
    const newTimestamp = moment(focusedTimestamp).subtract(dragDuration);
    this.jumpTo(newTimestamp);
  }

  handleZoom(e) {
    const scale = Math.pow(1.0015, e.deltaY);
    let durationPerPixel = scaleDuration(this.state.durationPerPixel, scale);
    if (durationPerPixel > MAX_DURATION_PER_PX) durationPerPixel = MAX_DURATION_PER_PX;
    if (durationPerPixel < MIN_DURATION_PER_PX) durationPerPixel = MIN_DURATION_PER_PX;
    this.setState({ durationPerPixel });
  }

  jumpTo(timestamp) {
    const { timestampNow } = this.state;
    const focusedTimestamp = timestamp > timestampNow ? timestampNow : timestamp;
    this.props.onUpdateTimestamp(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  jumpForward() {
    const { focusedTimestamp, durationPerPixel, boundingRect } = this.state;
    const duration = scaleDuration(durationPerPixel, boundingRect.width / 4);
    const timestamp = moment(focusedTimestamp).add(duration);
    this.jumpTo(timestamp);
  }

  jumpBackward() {
    const { focusedTimestamp, durationPerPixel, boundingRect } = this.state;
    const duration = scaleDuration(durationPerPixel, boundingRect.width / 4);
    const timestamp = moment(focusedTimestamp).subtract(duration);
    this.jumpTo(timestamp);
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  getTimeScale() {
    const { durationPerPixel, focusedTimestamp } = this.state;
    const roundedTimestamp = moment(focusedTimestamp).startOf('second').utc();
    const startDate = moment(roundedTimestamp).subtract(durationPerPixel);
    const endDate = moment(roundedTimestamp).add(durationPerPixel);
    return scaleUtc()
      .domain([startDate, endDate])
      .range([-1, 1]);
  }

  findOptimalDuration(durations) {
    const { durationPerPixel } = this.state;
    const minimalDuration = scaleDuration(durationPerPixel, MIN_TICK_SPACING_PX);
    return find(durations, d => d > minimalDuration);
  }

  renderTimestampTick({ timestamp, position, isBehind }, periodFormat, opacity) {
    const { timestampNow } = this.state;
    const disabled = timestamp.isAfter(timestampNow) || opacity < 0.2;
    const handleClick = () => this.jumpTo(timestamp);

    return (
      <g transform={`translate(${position}, 0)`} key={timestamp.format()}>
        {!isBehind && <line y2="75" stroke="#ddd" strokeWidth="1" />}
        <foreignObject width="100" height="20">
          <a className="timestamp-label" disabled={disabled} onClick={!disabled && handleClick}>
            {isBehind && '←'}
            {timestamp.utc().format(periodFormat)}
          </a>
        </foreignObject>
      </g>
    );
  }

  getTicks(period, parentPeriod, duration) {
    parentPeriod = parentPeriod || period;
    const startPosition = -this.state.boundingRect.width / 2;
    const endPosition = this.state.boundingRect.width / 2;

    if (!duration) return [];

    const timeScale = this.getTimeScale();
    const startDate = moment(timeScale.invert(startPosition));
    const endDate = moment(timeScale.invert(endPosition));
    const ticks = [];

    let timestamp = moment(startDate).startOf(parentPeriod);
    let turningPoint = moment(timestamp).add(1, parentPeriod);

    while (timestamp.isBefore(startDate)) {
      timestamp = moment(timestamp).add(duration);
    }

    ticks.push({
      timestamp: moment(timestamp).subtract(duration),
      position: startPosition,
      isBehind: true,
    });

    do {
      const position = timeScale(timestamp);

      while (ticks.length > 0 && position - last(ticks).position < 0.85 * MIN_TICK_SPACING_PX) {
        ticks.pop();
      }
      ticks.push({ timestamp, position });

      timestamp = moment(timestamp).add(duration);
      if (parentPeriod && timestamp >= turningPoint) {
        timestamp = turningPoint;
        turningPoint = moment(turningPoint).add(1, parentPeriod);
      }
    } while (timestamp.isBefore(endDate));

    return ticks;
  }

  renderPeriodBar(period, parentPeriod) {
    const duration = this.findOptimalDuration(TICK_ROWS[period].intervals);
    const ticks = this.getTicks(period, parentPeriod, duration);

    const periodFormat = TICK_ROWS[period].format;
    const p = getFadeInFactor(period, this.state.durationPerPixel);
    const verticalPosition = 60 - (p * 15);
    const opacity = Math.min(p, 1);

    return (
      <g className={period} transform={`translate(0, ${verticalPosition})`} style={{ opacity }}>
        {map(ticks, tick => this.renderTimestampTick(tick, periodFormat, opacity))}
      </g>
    );
  }

  renderDisabledShadow() {
    const timeScale = this.getTimeScale();
    const nowShift = timeScale(this.state.timestampNow);
    const { width, height } = this.state.boundingRect;

    return (
      <rect
        className="available-range"
        transform={`translate(${nowShift}, 0)`}
        width={width} height={height}
      />
    );
  }

  renderAxis() {
    return (
      <g id="axis">
        {this.renderDisabledShadow()}
        <g className="ticks">
          {this.renderPeriodBar('year')}
          {this.renderPeriodBar('month', 'year')}
          {this.renderPeriodBar('day', 'month')}
          {this.renderPeriodBar('minute', 'day')}
        </g>
      </g>
    );
  }

  render() {
    const className = classNames({ dragging: this.state.isPanning });
    const halfWidth = this.state.boundingRect.width / 2;

    return (
      <div className="time-travel-timeline">
        <a className="button jump-backward" onClick={this.jumpBackward}>
          <span className="fa fa-chevron-left" />
        </a>
        <svg className={className} ref={this.saveSvgRef} onWheel={this.handleZoom}>
          <g className="view" transform={`translate(${halfWidth}, 0)`}>
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
