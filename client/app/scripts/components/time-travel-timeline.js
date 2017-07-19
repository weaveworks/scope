import React from 'react';
import moment from 'moment';
import { debounce } from 'lodash';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import { zoom } from 'd3-zoom';
import { drag } from 'd3-drag';
// import { zoom, zoomIdentity } from 'd3-zoom';
import { scaleUtc } from 'd3-scale';
import { timeFormat } from 'd3-time-format';
import { timeMinute, timeHour, timeDay, timeMonth, timeYear } from 'd3-time';
import { event as d3Event, select } from 'd3-selection';
import {
  jumpToTime,
} from '../actions/app-actions';


import {
  TIMELINE_DEBOUNCE_INTERVAL,
} from '../constants/timer';

// import { transformToString } from '../utils/transform-utils';

const formatSecond = timeFormat(':%S');
const formatMinute = timeFormat('%I:%M');
const formatHour = timeFormat('%I %p');
const formatDay = timeFormat('%b %d');
const formatMonth = timeFormat('%B');
const formatYear = timeFormat('%Y');

function multiFormat(date) {
  date = new Date(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate(),
    date.getUTCHours(), date.getUTCMinutes(), date.getUTCSeconds());
  if (timeMinute(date) < date) return formatSecond(date);
  if (timeHour(date) < date) return formatMinute(date);
  if (timeDay(date) < date) return formatHour(date);
  if (timeMonth(date) < date) return formatDay(date);
  if (timeYear(date) < date) return formatMonth(date);
  return formatYear(date);
}


// const timeScale = scaleUtc().clamp(false)
//   .domain([new Date(1990, 1), new Date(2020, 1)])
//   .range([0, 10000]);

// const EARLIEST_TIMESTAMP = moment(new Date(2000, 0));
// const M = 1000;

class TimeTravelTimeline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      // shift: timeScale(moment().toDate()),
      translateX: 0,
      translateY: 15,
      scaleX: 1,
    };

    this.width = 2000;

    // this.getDisplayedTimeScale = this.getDisplayedTimeScale.bind(this);
    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.dragged = this.dragged.bind(this);
    this.zoomed = this.zoomed.bind(this);

    // this.zoomToDate

    this.debouncedUpdateTimestamp = debounce(
      this.updateTimestamp.bind(this), TIMELINE_DEBOUNCE_INTERVAL);
  }

  componentDidMount() {
    this.svg = select('svg#time-travel-timeline');
    // const w = 1200;
    // const h = 60;
    this.drag = drag()
      .on('drag', this.dragged);
    this.zoom = zoom()
      // .translateExtent([[0, 0], [w, h]])
      // .extent([[-M, -30], [M, 30]])
      .on('zoom', this.zoomed);
    // this.svg.style('transform-origin', '50% 50% 0');
    // console.log(this.svg.getBoundingClientRect());

    this.setZoomTriggers(true);
    // this.updateZoomLimits(this.props);
    // this.restoreZoomState(this.props);
  }

  componentWillUnmount() {
    this.setZoomTriggers(false);
  }

  componentWillReceiveProps() {
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

  getTimeScale() {
    const { scaleX, translateX } = this.state;
    const S = 4000 / scaleX;
    const p = this.props.pausedAt ? moment(this.props.pausedAt) : moment();
    const x = p.add(translateX / S, 'hours');
    const k = S / scaleX;
    console.log(x, k);
    // const x = moment(p).subtract(this.state.translateX, 'seconds');
    // // console.log('L', p, x);
    // // const pausedAt = this.props.pausedAt || moment();
    // const k = 1000 / this.state.scaleX;
    // return scaleUtc()
    //   .domain([x.subtract(k, 'seconds').toDate(), x.add(k, 'seconds').toDate()])
    //   .range([0, 2000]);
    return scaleUtc()
      .domain([x.subtract(k, 'hours').toDate(), x.add(k, 'hours').toDate()])
      .range([-2000 / scaleX, 2000 / scaleX]);
  }

  zoomed() {
    // const { x, k } = d3Event.transform;
    // const { scaleX, translateX } = this.state;
    // const { width } = this.svgRef.getBoundingClientRect();
    // // const halfWidth = 0; // width / 2;
    // const diff = (width * 0.5) - translateX;
    // const fac = diff * 0.5 * (1 - (k / scaleX));
    // console.log(diff, fac);
    //
    // if (k !== scaleX) {
    //   const newTransform = zoomIdentity
    //     .translate(translateX, 0)
    //     .scale(k)
    //     .translate(fac, 0);
    //   console.log(newTransform);
    //   this.setState({ translateX: newTransform.x, scaleX: newTransform.k });
    //   this.svg.call(this.zoom.transform, newTransform);
    // } else if (x !== translateX) {
    //   const newTransform = zoomIdentity
    //     .translate(x, 0)
    //     .scale(scaleX);
    //   console.log(newTransform);
    //   this.setState({ translateX: newTransform.x, scaleX: newTransform.k });
    //   this.svg.call(this.zoom.transform, newTransform);
    // }
    // this.svg.call(this.zoom.transform, zoomIdentity
    //   .translate(this.state.translateX, 0)
    //   .scale(this.state.scaleX, 1));
    // console.log(d3Event);
    // console.log(zoomTransform(this.svgRef));
    // console.log('ZOOM', d3Event.transform.k);
    this.setState({ scaleX: d3Event.transform.k });

    // this.debouncedUpdateTimestamp(this.getDisplayedTimeScale().invert(-d3Event.transform.x));
  }

  dragged() {
    // console.log('DRAG', this.state.translateX + d3Event.dx);
    this.setState({ translateX: this.state.translateX + d3Event.dx });
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  renderAxis() {
    // const { translateX, scaleX } = this.state;
    // const pausedAt = this.props.pausedAt ? moment(this.props.pausedAt) : moment();
    // const timeScale = this.getDisplayedTimeScale();
    //
    // const startDate = this.getDisplayedTimeScale()
    //   .invert(-this.state.translateX - (1000 * this.state.scaleX));
    // const endDate = this.getDisplayedTimeScale()
    //   .invert(-this.state.translateX + (1000 * this.state.scaleX));
    // const ticks = this.getDisplayedTimeScale().domain([startDate, endDate]).ticks(10);
    // console.log(startDate, -this.state.translateX - (1000 * this.state.scaleX));
    // console.log(endDate, -this.state.translateX + (1000 * this.state.scaleX));
    // console.log(ticks);

    const timeScale = this.getTimeScale();
    // const cd = pausedAt.subtract((translateX / scaleX) * (30 / 10000), 'years');
    // console.log(cd.toDate());
    const ticks = timeScale.ticks(10);

    // <line x1="-1000"x2="1000" stroke="#ddd" strokeWidth="1" />
    return (
      <g id="axis">
        <g className="ticks">
          {fromJS(ticks).map(date => (
            <foreignObject
              transform={`translate(${timeScale(date)}, 0) scale(${1 / this.state.scaleX})`}
              key={moment(date).format()}
              width="50" height="20">
              <span>{multiFormat(date)}</span>
            </foreignObject>
          ))}
        </g>
      </g>
    );
  }

  render() {
    const { translateX, scaleX } = this.state;
    return (
      <div className="time-travel-timeline">
        <a className="button jump-backward" onClick={this.jumpBackward}>
          <span className="fa fa-chevron-left" />
        </a>
        <svg
          viewBox={`${-this.width / 2} -30 ${this.width} 60`}
          id="time-travel-timeline"
          width="100%" height="100%"
          ref={this.saveSvgRef}>
          <g className="view" transform={`translate(${translateX}, 0) scale(${scaleX})`}>
            {this.renderAxis()}
          </g>
        </svg>
        <a className="button jump-forward" onClick={this.jumpBackward}>
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
