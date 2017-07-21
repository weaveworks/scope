import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { debounce } from 'lodash';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import { zoom } from 'd3-zoom';
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


const fixedDurations = [
  moment.duration(1, 'minute'),
  moment.duration(5, 'minutes'),
  moment.duration(15, 'minutes'),
  moment.duration(1, 'hour'),
  moment.duration(3, 'hours'),
  moment.duration(6, 'hours'),
  moment.duration(1, 'day'),
  moment.duration(1, 'week'),
  moment.duration(1, 'month'),
  moment.duration(3, 'months'),
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
  console.log(scale, yearShift, monthShift, dayShift, minuteShift);
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

const R = 2000;
const C = 1000000;

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
    this.dragStarted = this.dragStarted.bind(this);
    this.dragEnded = this.dragEnded.bind(this);
    this.dragged = this.dragged.bind(this);
    this.zoomed = this.zoomed.bind(this);
    this.jumpTo = this.jumpTo.bind(this);
    this.jumpForward = this.jumpForward.bind(this);
    this.jumpBackward = this.jumpBackward.bind(this);

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    this.svg = select('svg#time-travel-timeline');
    this.drag = drag()
      .on('start', this.dragStarted)
      .on('end', this.dragEnded)
      .on('drag', this.dragged);
    this.zoom = zoom()
      .scaleExtent([0.003, 1000])
      .on('zoom', this.zoomed);

    this.setZoomTriggers(true);

    // Force periodic re-renders to update the slider position as time goes by.
    this.timer = setInterval(() => {
      this.setState({ timestampNow: moment().startOf('second') });
    }, TIMELINE_TICK_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
    this.setZoomTriggers(false);
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

  setZoomTriggers(zoomingEnabled) {
    if (zoomingEnabled) {
      this.svg.call(this.drag);
      // use d3-zoom defaults but exclude double clicks
      this.svg.call(this.zoom)
        .on('dblclick.zoom', null);
    } else {
      this.svg.on('.zoom', null);
    }
  }

  zoomed() {
    const timelineRange = moment.duration(C / d3Event.transform.k, 'seconds');
    // console.log('ZOOM', timelineRange.toJSON());
    this.setState({ timelineRange, scaleX: d3Event.transform.k });
  }

  dragStarted() {
    this.setState({ isDragging: true });
  }

  dragged() {
    const { focusedTimestamp, timelineRange } = this.state;
    const mv = timelineRange.asMilliseconds() / R;
    const newTimestamp = moment(focusedTimestamp).subtract(d3Event.dx * mv);
    // console.log('DRAG', newTimestamp.toDate());
    this.jumpTo(newTimestamp);
  }

  dragEnded() {
    this.setState({ isDragging: false });
  }

  jumpTo(timestamp) {
    const { timestampNow } = this.state;
    const focusedTimestamp = timestamp > timestampNow ? timestampNow : timestamp;
    this.props.onUpdateTimestamp(focusedTimestamp);
    this.setState({ focusedTimestamp });
  }

  jumpForward() {
    const d = this.state.timelineRange.asMilliseconds() / 4 / R;
    const timestamp = moment(this.state.focusedTimestamp).add(d * this.width);
    this.jumpTo(timestamp);
  }

  jumpBackward() {
    const d = this.state.timelineRange.asMilliseconds() / 4 / R;
    const timestamp = moment(this.state.focusedTimestamp).subtract(d * this.width);
    this.jumpTo(timestamp);
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  getTimeScale() {
    const { timelineRange, focusedTimestamp } = this.state;
    const rTimestamp = moment(focusedTimestamp).startOf('second').utc();
    const startDate = moment(rTimestamp).subtract(timelineRange);
    const endDate = moment(rTimestamp).add(timelineRange);
    return scaleUtc()
      .domain([startDate, endDate])
      .range([-R, R]);
  }

  renderPeriodBar(period, prevPeriod, periodFormat, [startIndex, endIndex]) {
    const timeScale = this.getTimeScale();
    const startDate = moment(timeScale.invert(-R));
    const endDate = moment(timeScale.invert(R));
    const numSeconds = endDate.diff(startDate, 'seconds', true);
    // console.log(numSeconds);
    const ts = [];

    let duration = null;
    for (let i = startIndex; i <= endIndex; i += 1) {
      if (numSeconds / fixedDurations[i].asSeconds() < 50) {
        duration = fixedDurations[i];
        break;
      }
    }

    const behind = (period === 'day') ? 2 : 0;

    // console.log(duration.asSeconds());
    if (!duration) return null;

    let t = moment(startDate).startOf(prevPeriod);
    let turningPoint = moment(t).add(1, prevPeriod);
    do {
      const p = timeScale(t);
      if (p > -this.width && p < this.width) {
        ts.push(t);
      }
      t = moment(t).add(duration);
      if (prevPeriod !== period && t >= moment(turningPoint).subtract(behind, period)) {
        t = turningPoint;
        turningPoint = moment(turningPoint).add(1, prevPeriod);
      }
    } while (timeScale(t) < this.width);

    // console.log(ts);

    const p = getShift(period, this.state.scaleX);
    const shift = 60 * (1 - (p * 0.25));
    const opacity = Math.min(p * p, 1);
    return (
      <g className={period} transform={`translate(0, ${shift})`} style={{ opacity }}>
        {fromJS(ts).map(timestamp => (
          <g transform={`translate(${timeScale(timestamp)}, 0)`} key={timestamp.format()}>
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

  renderAxis() {
    const timeScale = this.getTimeScale();
    const nowX = Math.min(timeScale(this.state.timestampNow), R);

    return (
      <g id="axis">
        <rect
          className="available-range"
          transform={`translate(${nowX}, 0)`}
          x={-2 * R} y={0} width={2 * R} height={70} />
        <g className="ticks">
          {this.renderPeriodBar('year', 'year', 'YYYY', [10, 10])}
          {this.renderPeriodBar('month', 'year', 'MMMM', [8, 9])}
          {this.renderPeriodBar('day', 'month', 'Do', [6, 7])}
          {this.renderPeriodBar('minute', 'day', 'HH:mm', [0, 5])}
        </g>
      </g>
    );
  }

  render() {
    const className = classNames({ dragging: this.state.isDragging });
    return (
      <div className="time-travel-timeline">
        <a className="button jump-backward" onClick={this.jumpBackward}>
          <span className="fa fa-chevron-left" />
        </a>
        <svg
          className={className}
          id="time-travel-timeline"
          viewBox={`${-this.width / 2} 0 ${this.width} 70`}
          width="100%" height="100%"
          ref={this.saveSvgRef}>
          <g className="view">
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
