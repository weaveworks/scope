import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { map, clamp, find, last } from 'lodash';
import { connect } from 'react-redux';
import { drag } from 'd3-drag';
import { scaleUtc } from 'd3-scale';
import { event as d3Event, select } from 'd3-selection';

import {
  nowInSecondsPrecision,
  clampToNowInSecondsPrecision,
  scaleDuration,
} from '../utils/time-utils';

import { TIMELINE_TICK_INTERVAL } from '../constants/timer';


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

const MIN_DURATION_PER_PX = moment.duration(250, 'milliseconds');
const INIT_DURATION_PER_PX = moment.duration(1, 'minute');
const MAX_DURATION_PER_PX = moment.duration(3, 'days');
const MIN_TICK_SPACING_PX = 80;
const MAX_TICK_SPACING_PX = 415;
const ZOOM_SENSITIVITY = 1.0015;
const FADE_OUT_FACTOR = 1.4;


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
      timestampNow: nowInSecondsPrecision(),
      focusedTimestamp: nowInSecondsPrecision(),
      durationPerPixel: INIT_DURATION_PER_PX,
      boundingRect: { width: 0, height: 0 },
      isDragging: false,
    };

    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.jumpRelativePixels = this.jumpRelativePixels.bind(this);
    this.jumpForward = this.jumpForward.bind(this);
    this.jumpBackward = this.jumpBackward.bind(this);
    this.jumpTo = this.jumpTo.bind(this);

    this.findOptimalDuration = this.findOptimalDuration.bind(this);

    this.handleZoom = this.handleZoom.bind(this);
    this.handlePanStart = this.handlePanStart.bind(this);
    this.handlePanEnd = this.handlePanEnd.bind(this);
    this.handlePan = this.handlePan.bind(this);
  }

  componentDidMount() {
    this.svg = select('.time-travel-timeline svg');
    this.drag = drag()
      .on('start', this.handlePanStart)
      .on('end', this.handlePanEnd)
      .on('drag', this.handlePan);
    this.svg.call(this.drag);

    // Force periodic updates of the availability range as time goes by.
    this.timer = setInterval(() => {
      this.setState({ timestampNow: nowInSecondsPrecision() });
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

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.props.onTimelinePanEnd(this.state.focusedTimestamp);
    this.setState({ isPanning: false });
  }

  handlePan() {
    const dragDuration = scaleDuration(this.state.durationPerPixel, -d3Event.dx);
    const timestamp = moment(this.state.focusedTimestamp).add(dragDuration);
    const focusedTimestamp = clampToNowInSecondsPrecision(timestamp);
    this.props.onTimelinePan(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  handleZoom(e) {
    const scale = Math.pow(ZOOM_SENSITIVITY, e.deltaY);
    let durationPerPixel = scaleDuration(this.state.durationPerPixel, scale);
    if (durationPerPixel > MAX_DURATION_PER_PX) durationPerPixel = MAX_DURATION_PER_PX;
    if (durationPerPixel < MIN_DURATION_PER_PX) durationPerPixel = MIN_DURATION_PER_PX;
    this.setState({ durationPerPixel });
  }

  jumpTo(timestamp) {
    const focusedTimestamp = clampToNowInSecondsPrecision(timestamp);
    this.props.onInstantJump(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  jumpRelativePixels(pixels) {
    const duration = scaleDuration(this.state.durationPerPixel, pixels);
    const timestamp = moment(this.state.focusedTimestamp).add(duration);
    this.jumpTo(timestamp);
  }

  jumpForward() {
    this.jumpRelativePixels(this.state.boundingRect.width / 4);
  }

  jumpBackward() {
    this.jumpRelativePixels(-this.state.boundingRect.width / 4);
  }

  getTimeScale() {
    const { durationPerPixel, focusedTimestamp } = this.state;
    const roundedTimestamp = moment(focusedTimestamp).utc().startOf('second');
    const startDate = moment(roundedTimestamp).subtract(durationPerPixel);
    const endDate = moment(roundedTimestamp).add(durationPerPixel);
    return scaleUtc()
      .domain([startDate, endDate])
      .range([-1, 1]);
  }

  findOptimalDuration(durations) {
    const { durationPerPixel } = this.state;
    const minimalDuration = scaleDuration(durationPerPixel, 1.1 * MIN_TICK_SPACING_PX);
    return find(durations, d => d >= minimalDuration);
  }

  getTicks(period, parentPeriod) {
    // First find the optimal duration between the ticks - if no satisfactory
    // duration could be found, don't render any ticks for the given period.
    const duration = this.findOptimalDuration(TICK_ROWS[period].intervals);
    if (!duration) return [];

    // Get the boundary values for the displayed part of the timeline.
    const timeScale = this.getTimeScale();
    const startPosition = -this.state.boundingRect.width / 2;
    const endPosition = this.state.boundingRect.width / 2;
    const startDate = moment(timeScale.invert(startPosition));
    const endDate = moment(timeScale.invert(endPosition));

    // Start counting the timestamps from the most recent timestamp that is not shown
    // on screen. The values are always rounded up to the timestamps of the next bigger
    // period (e.g. for days it would be months, for months it would be years).
    let timestamp = moment(startDate).utc().startOf(parentPeriod || period);
    while (timestamp.isBefore(startDate)) {
      timestamp = moment(timestamp).add(duration);
    }
    timestamp = moment(timestamp).subtract(duration);

    // Make that hidden timestamp the first one in the list, but position
    // it inside the visible range with a prepended arrow to the past.
    const ticks = [{
      timestamp: moment(timestamp),
      position: startPosition,
      isBehind: true,
    }];

    // Continue adding ticks till the end of the visible range.
    do {
      // If the new timestamp enters into a new bigger period, we round it down to the
      // beginning of that period. E.g. instead of going [Jan 22nd, Jan 29th, Feb 5th],
      // we output [Jan 22nd, Jan 29th, Feb 1st]. Right now this case only happens between
      // days and months, but in theory it could happen whenever bigger periods are not
      // divisible by the duration we are using as a step between the ticks.
      let newTimestamp = moment(timestamp).add(duration);
      if (parentPeriod && newTimestamp.get(parentPeriod) !== timestamp.get(parentPeriod)) {
        newTimestamp = moment(newTimestamp).utc().startOf(parentPeriod);
      }
      timestamp = newTimestamp;

      // If the new tick is too close to the previous one, drop that previous tick.
      const position = timeScale(timestamp);
      const previousPosition = last(ticks) && last(ticks).position;
      if (position - previousPosition < MIN_TICK_SPACING_PX) {
        ticks.pop();
      }

      ticks.push({ timestamp, position });
    } while (timestamp.isBefore(endDate));

    return ticks;
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
            {isBehind && '←'}{timestamp.utc().format(periodFormat)}
          </a>
        </foreignObject>
      </g>
    );
  }

  renderPeriodBar(period, parentPeriod) {
    const ticks = this.getTicks(period, parentPeriod);

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
    const className = classNames({ panning: this.state.isPanning });
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

export default connect(mapStateToProps)(TimeTravelTimeline);
