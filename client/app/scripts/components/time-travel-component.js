import React from 'react';
import moment from 'moment';
import styled from 'styled-components';
import { map, clamp, find, last, debounce } from 'lodash';
import { drag } from 'd3-drag';
import { scaleUtc } from 'd3-scale';
import { event as d3Event, select } from 'd3-selection';
import { Motion } from 'react-motion';

import { zoomFactor } from '../utils/zoom-utils';
import { strongSpring } from '../utils/animation-utils';
import { linearGradientValue } from '../utils/math-utils';
import {
  nowInSecondsPrecision,
  clampToNowInSecondsPrecision,
  scaleDuration,
} from '../utils/time-utils';


const TICK_SETTINGS_PER_PERIOD = {
  year: {
    format: 'YYYY',
    childPeriod: 'month',
    intervals: [
      moment.duration(1, 'year'),
    ],
  },
  month: {
    format: 'MMMM',
    parentPeriod: 'year',
    childPeriod: 'day',
    intervals: [
      moment.duration(1, 'month'),
      moment.duration(3, 'months'),
    ],
  },
  day: {
    format: 'Do',
    parentPeriod: 'month',
    childPeriod: 'minute',
    intervals: [
      moment.duration(1, 'day'),
      moment.duration(1, 'week'),
    ],
  },
  minute: {
    format: 'HH:mm',
    parentPeriod: 'day',
    intervals: [
      moment.duration(1, 'minute'),
      moment.duration(5, 'minutes'),
      moment.duration(15, 'minutes'),
      moment.duration(1, 'hour'),
      moment.duration(3, 'hours'),
      moment.duration(6, 'hours'),
    ],
  },
};

const ZOOM_TRACK_DEBOUNCE_INTERVAL = 5000;
const TIMELINE_DEBOUNCE_INTERVAL = 500;
const TIMELINE_TICK_INTERVAL = 1000;

const TIMELINE_HEIGHT = '55px';
const MIN_DURATION_PER_PX = moment.duration(250, 'milliseconds');
const INIT_DURATION_PER_PX = moment.duration(1, 'minute');
const MAX_DURATION_PER_PX = moment.duration(3, 'days');
const MIN_TICK_SPACING_PX = 70;
const MAX_TICK_SPACING_PX = 415;
const FADE_OUT_FACTOR = 1.4;
const TICKS_ROW_SPACING = 16;
const MAX_TICK_ROWS = 3;


// From https://stackoverflow.com/a/18294634
const FullyPannableCanvas = styled.svg`
  width: 100%;
  height: 100%;
  cursor: move;
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;

  ${props => props.panning && `
    cursor: grabbing;
    cursor: -moz-grabbing;
    cursor: -webkit-grabbing;
  `}
`;

const TimeTravelContainer = styled.div`
  transition: all .15s ease-in-out;
  position: relative;
  margin-bottom: 15px;
  overflow: hidden;
  z-index: 2001;
  height: 0;

  ${props => props.visible && `
    height: calc(${TIMELINE_HEIGHT} + 35px);
    margin-bottom: 15px;
    margin-top: -5px;
  `}
`;

const TimelineContainer = styled.div`
  align-items: center;
  display: flex;
  height: ${TIMELINE_HEIGHT};

  &:before, &:after {
    border: 1px solid ${props => props.theme.colors.white};
    background-color: ${props => props.theme.colors.accent.orange};
    content: '';
    position: absolute;
    display: block;
    left: 50%;
    border-top: 0;
    border-bottom: 0;
    margin-left: -1px;
    width: 3px;
  }

  &:before {
    top: 0;
    height: ${TIMELINE_HEIGHT};
  }

  &:after {
    top: ${TIMELINE_HEIGHT};
    height: 9px;
    opacity: 0.15;
  }
`;

const Timeline = FullyPannableCanvas.extend`
  background-color: rgba(255, 255, 255, 0.85);
  box-shadow: inset 0 0 7px ${props => props.theme.colors.gray};
  pointer-events: all;
  margin: 0 7px;
`;

const DisabledRange = styled.rect`
  fill: ${props => props.theme.colors.gray};
  fill-opacity: 0.15;
`;

const TimestampLabel = styled.a`
  margin-left: 2px;
  padding: 3px;

  &[disabled] {
    color: ${props => props.theme.colors.gray};
    cursor: inherit;
  }
`;

const TimelinePanButton = styled.a`
  pointer-events: all;
  padding: 2px;
`;

const TimestampContainer = styled.div`
  background-color: ${props => props.theme.colors.white};
  border: 1px solid ${props => props.theme.colors.gray};
  border-radius: 4px;
  padding: 2px 8px;
  pointer-events: all;
  margin: 8px 0 25px 50%;
  transform: translateX(-50%);
  opacity: 0.8;
  display: inline-block;
`;

const TimestampInput = styled.input`
  background-color: transparent;
  font-family: "Roboto", sans-serif;
  text-align: center;
  font-size: 1rem;
  width: 165px;
  border: 0;
  outline: 0;
`;


function getTimeScale({ focusedTimestamp, durationPerPixel }) {
  const roundedTimestamp = moment(focusedTimestamp).utc().startOf('second');
  const startDate = moment(roundedTimestamp).subtract(durationPerPixel);
  const endDate = moment(roundedTimestamp).add(durationPerPixel);
  return scaleUtc()
    .domain([startDate, endDate])
    .range([-1, 1]);
}

function findOptimalDurationFit(durations, { durationPerPixel }) {
  const minimalDuration = scaleDuration(durationPerPixel, 1.1 * MIN_TICK_SPACING_PX);
  return find(durations, d => d >= minimalDuration);
}

function getInputValue(timestamp) {
  return {
    inputValue: (timestamp ? moment(timestamp) : moment()).utc().format(),
  };
}

export default class TimeTravelComponent extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      timestampNow: nowInSecondsPrecision(),
      focusedTimestamp: nowInSecondsPrecision(),
      durationPerPixel: INIT_DURATION_PER_PX,
      boundingRect: { width: 0, height: 0 },
      isPanning: false,
      ...getInputValue(props.timestamp),
    };

    this.jumpRelativePixels = this.jumpRelativePixels.bind(this);
    this.jumpForward = this.jumpForward.bind(this);
    this.jumpBackward = this.jumpBackward.bind(this);
    this.jumpTo = this.jumpTo.bind(this);

    this.handleZoom = this.handleZoom.bind(this);
    this.handlePanStart = this.handlePanStart.bind(this);
    this.handlePanEnd = this.handlePanEnd.bind(this);
    this.handlePan = this.handlePan.bind(this);

    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.debouncedTrackZoom = debounce(this.trackZoom.bind(this), ZOOM_TRACK_DEBOUNCE_INTERVAL);

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleTimelinePan = this.handleTimelinePan.bind(this);
    this.handleTimelinePanEnd = this.handleTimelinePanEnd.bind(this);
    this.handleInstantJump = this.handleInstantJump.bind(this);

    this.instantUpdateTimestamp = this.instantUpdateTimestamp.bind(this);
    this.debouncedUpdateTimestamp = debounce(
      this.instantUpdateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
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
    // Update the input value
    this.setState(getInputValue(nextProps.timestamp));
    // Don't update the focused timestamp if we're not paused (so the timeline is hidden).
    if (nextProps.timestamp) {
      this.setState({ focusedTimestamp: nextProps.timestamp });
    }
    // Always update the timeline dimension information.
    this.setState({ boundingRect: this.svgRef.getBoundingClientRect() });
  }

  handleInputChange(ev) {
    const timestamp = moment(ev.target.value);
    this.setState({ inputValue: ev.target.value });

    if (timestamp.isValid()) {
      const clampedTimestamp = clampToNowInSecondsPrecision(timestamp);
      this.instantUpdateTimestamp(clampedTimestamp, this.props.trackTimestampEdit);
    }
  }

  handleTimelinePan(timestamp) {
    this.setState(getInputValue(timestamp));
    this.debouncedUpdateTimestamp(timestamp);
  }

  handleTimelinePanEnd(timestamp) {
    this.instantUpdateTimestamp(timestamp, this.props.trackTimelinePan);
  }

  handleInstantJump(timestamp) {
    this.instantUpdateTimestamp(timestamp, this.props.trackTimelineClick);
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.handleTimelinePanEnd(this.state.focusedTimestamp);
    this.setState({ isPanning: false });
  }

  handlePan() {
    const dragDuration = scaleDuration(this.state.durationPerPixel, -d3Event.dx);
    const timestamp = moment(this.state.focusedTimestamp).add(dragDuration);
    const focusedTimestamp = clampToNowInSecondsPrecision(timestamp);
    this.handleTimelinePan(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  handleZoom(ev) {
    let durationPerPixel = scaleDuration(this.state.durationPerPixel, 1 / zoomFactor(ev));
    if (durationPerPixel > MAX_DURATION_PER_PX) durationPerPixel = MAX_DURATION_PER_PX;
    if (durationPerPixel < MIN_DURATION_PER_PX) durationPerPixel = MIN_DURATION_PER_PX;

    this.setState({ durationPerPixel });
    this.debouncedTrackZoom();
    ev.preventDefault();
  }

  instantUpdateTimestamp(timestamp, callback) {
    if (!timestamp.isSame(this.props.timestamp)) {
      this.debouncedUpdateTimestamp.cancel();
      this.setState(getInputValue(timestamp));
      this.props.changeTimestamp(moment(timestamp));

      // Used for tracking.
      if (callback) callback();
    }
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  trackZoom() {
    const periods = ['years', 'months', 'weeks', 'days', 'hours', 'minutes', 'seconds'];
    const duration = scaleDuration(this.state.durationPerPixel, MAX_TICK_SPACING_PX);
    const zoomedPeriod = find(periods, period => Math.floor(duration.get(period)) && period);
    this.props.trackTimelineZoom(zoomedPeriod);
  }

  jumpTo(timestamp) {
    const focusedTimestamp = clampToNowInSecondsPrecision(timestamp);
    this.handleInstantJump(focusedTimestamp);
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

  getVerticalShiftForPeriod(period, { durationPerPixel }) {
    const { childPeriod, parentPeriod } = TICK_SETTINGS_PER_PERIOD[period];

    let shift = 1;
    if (parentPeriod) {
      const durationMultiplier = 1 / MAX_TICK_SPACING_PX;
      const parentPeriodStartInterval = TICK_SETTINGS_PER_PERIOD[parentPeriod].intervals[0];
      const fadedInDuration = scaleDuration(parentPeriodStartInterval, durationMultiplier);
      const fadedOutDuration = scaleDuration(fadedInDuration, FADE_OUT_FACTOR);

      const durationLog = d => Math.log(d.asMilliseconds());
      const transitionFactor = durationLog(fadedOutDuration) - durationLog(durationPerPixel);
      const transitionLength = durationLog(fadedOutDuration) - durationLog(fadedInDuration);

      shift = clamp(transitionFactor / transitionLength, 0, 1);
    }

    if (childPeriod) {
      shift += this.getVerticalShiftForPeriod(childPeriod, { durationPerPixel });
    }

    return shift;
  }

  getTicksForPeriod(period, timelineTransform) {
    // First find the optimal duration between the ticks - if no satisfactory
    // duration could be found, don't render any ticks for the given period.
    const { parentPeriod, intervals } = TICK_SETTINGS_PER_PERIOD[period];
    const duration = findOptimalDurationFit(intervals, timelineTransform);
    if (!duration) return [];

    // Get the boundary values for the displayed part of the timeline.
    const timeScale = getTimeScale(timelineTransform);
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
    // Ticks are disabled if they are in the future or if they are too transparent.
    const disabled = timestamp.isAfter(this.state.timestampNow) || opacity < 0.4;
    const handleClick = () => this.jumpTo(timestamp);

    return (
      <g transform={`translate(${position}, 0)`} key={timestamp.format()}>
        {!isBehind && <line y2="75" stroke="#ddd" strokeWidth="1" />}
        {!disabled && <title>Jump to {timestamp.utc().format()}</title>}
        <foreignObject width="100" height="20" style={{ lineHeight: '20px' }}>
          <TimestampLabel disabled={disabled} onClick={!disabled && handleClick}>
            {timestamp.utc().format(periodFormat)}
          </TimestampLabel>
        </foreignObject>
      </g>
    );
  }

  renderPeriodTicks(period, timelineTransform) {
    const periodFormat = TICK_SETTINGS_PER_PERIOD[period].format;
    const ticks = this.getTicksForPeriod(period, timelineTransform);

    const ticksRow = MAX_TICK_ROWS - this.getVerticalShiftForPeriod(period, timelineTransform);
    const transform = `translate(0, ${ticksRow * TICKS_ROW_SPACING})`;

    // Ticks quickly fade in from the bottom and then slowly start
    // fading out towards the top until they are pushed out of canvas.
    const focusedRow = MAX_TICK_ROWS - 1;
    const opacity = ticksRow > focusedRow ?
      linearGradientValue(ticksRow, [MAX_TICK_ROWS, focusedRow]) :
      linearGradientValue(ticksRow, [-2, focusedRow]);

    return (
      <g className={period} transform={transform} style={{ opacity }}>
        {map(ticks, tick => this.renderTimestampTick(tick, periodFormat, opacity))}
      </g>
    );
  }

  renderDisabledShadow(timelineTransform) {
    const timeScale = getTimeScale(timelineTransform);
    const nowShift = timeScale(this.state.timestampNow);
    const { width, height } = this.state.boundingRect;

    return (
      <DisabledRange transform={`translate(${nowShift}, 0)`} width={width} height={height} />
    );
  }

  renderAxis(timelineTransform) {
    const { width, height } = this.state.boundingRect;

    return (
      <g id="axis">
        <rect
          className="tooltip-container"
          transform={`translate(${-width / 2}, 0)`}
          width={width} height={height} fillOpacity={0}
        />
        {this.renderDisabledShadow(timelineTransform)}
        <g className="ticks" transform="translate(0, 1)">
          {this.renderPeriodTicks('year', timelineTransform)}
          {this.renderPeriodTicks('month', timelineTransform)}
          {this.renderPeriodTicks('day', timelineTransform)}
          {this.renderPeriodTicks('minute', timelineTransform)}
        </g>
      </g>
    );
  }

  renderAnimatedContent() {
    const focusedTimestampValue = this.state.focusedTimestamp.valueOf();
    const durationPerPixelValue = this.state.durationPerPixel.asMilliseconds();

    return (
      <Motion
        style={{
          focusedTimestampValue: strongSpring(focusedTimestampValue),
          durationPerPixelValue: strongSpring(durationPerPixelValue),
        }}>
        {interpolated => this.renderAxis({
          focusedTimestamp: moment(interpolated.focusedTimestampValue),
          durationPerPixel: moment.duration(interpolated.durationPerPixelValue),
        })}
      </Motion>
    );
  }

  render() {
    const { isPanning, boundingRect } = this.state;
    const halfWidth = boundingRect.width / 2;

    return (
      <TimeTravelContainer visible={this.props.visible}>
        <TimelineContainer className="time-travel-timeline">
          <TimelinePanButton onClick={this.jumpBackward}>
            <span className="fa fa-chevron-left" />
          </TimelinePanButton>
          <Timeline panning={isPanning} innerRef={this.saveSvgRef} onWheel={this.handleZoom}>
            <g className="timeline-container" transform={`translate(${halfWidth}, 0)`}>
              <title>Scroll to zoom, drag to pan</title>
              {this.renderAnimatedContent()}
            </g>
          </Timeline>
          <TimelinePanButton onClick={this.jumpForward}>
            <span className="fa fa-chevron-right" />
          </TimelinePanButton>
        </TimelineContainer>
        <TimestampContainer>
          <TimestampInput
            value={this.state.inputValue}
            onChange={this.handleInputChange}
          /> UTC
        </TimestampContainer>
      </TimeTravelContainer>
    );
  }
}
