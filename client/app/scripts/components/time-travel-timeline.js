import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import { zoom } from 'd3-zoom';
import { scaleTime } from 'd3-scale';
// import { timeFormat } from 'd3-time-format';
// import { timeSecond, timeMinute, timeHour, timeDay, timeWeek, timeMonth } from 'd3-time';
import { event as d3Event, select } from 'd3-selection';

import { transformToString } from '../utils/transform-utils';


// function multiFormat(date) {
//   if (timeSecond(date) < date) timeFormat(':%S');
//   if (timeMinute(date) < date) timeFormat('%I:%M');
//   if (timeHour(date) < date) timeFormat('%I %p');
//   if (timeDay(date) < date) timeFormat('%a %d');
//   if (timeWeek(date) < date) timeFormat('%b %d');
//   if (timeMonth(date) < date) timeFormat('%B');
//   return timeFormat('%Y');
// }
// const customTimeFormat = d3.time.format.multi([
//   [".%L", function(d) { return d.getMilliseconds(); }],
//   [":%S", function(d) { return d.getSeconds(); }],
//   ["%I:%M", function(d) { return d.getMinutes(); }],
//   ["%I %p", function(d) { return d.getHours(); }],
//   ["%a %d", function(d) { return d.getDay() && d.getDate() != 1; }],
//   ["%b %d", function(d) { return d.getDate() != 1; }],
//   ["%B", function(d) { return d.getMonth(); }],
//   ["%Y", function() { return true; }]
// ]);

const EARLIEST_TIMESTAMP = moment(new Date(2000, 0));

class TimeTravelTimeline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      translateX: 0,
      translateY: 15,
      scaleX: 1,
    };

    this.x = scaleTime()
      .domain([EARLIEST_TIMESTAMP.toDate(), moment().toDate()])
      .range([-10000, 0]);

    this.saveSvgRef = this.saveSvgRef.bind(this);
    this.zoomed = this.zoomed.bind(this);
  }

  componentDidMount() {
    this.zoom = zoom().on('zoom', this.zoomed);
    this.svg = select('svg#time-travel-timeline');
    // console.log(this.svg.getBoundingClientRect());

    this.setZoomTriggers(true);
    // this.updateZoomLimits(this.props);
    // this.restoreZoomState(this.props);
  }

  componentWillUnmount() {
    this.setZoomTriggers(false);
  }

  componentWillReceiveProps() {
    console.log(this.svgRef.getBoundingClientRect());
  }

  setZoomTriggers(zoomingEnabled) {
    if (zoomingEnabled) {
      // use d3-zoom defaults but exclude double clicks
      this.svg.call(this.zoom)
        .on('dblclick.zoom', null);
    } else {
      this.svg.on('.zoom', null);
    }
  }

  zoomed() {
    this.setState({
      translateX: d3Event.transform.x,
      scaleX: d3Event.transform.k,
    });
  }

  saveSvgRef(ref) {
    this.svgRef = ref;
  }

  renderAxis() {
    const ticks = this.x.domain([new Date(2000, 0, 1, 0), new Date(2001, 0, 1, 0)]).ticks(10);
    return (
      <g id="axis">
        <line x1="0" y1="30" x2="1000" y2="30" stroke="#ddd" strokeWidth="1" />
        <g className="ticks" transform={`translate(${this.state.translateX}, 0)`}>
          {fromJS(ticks).map(date => (
            <rect
              key={moment(date).format()}
              width="20" height="20" />
          ))}
        </g>
      </g>
    );
  }

  render() {
    return (
      <svg id="time-travel-timeline" width="100%" height="100%" ref={this.saveSvgRef}>
        <g id="view" transform={transformToString(this.state)}>
          <rect x="10" y="10" width="100" height="30" />
        </g>
        {this.renderAxis()}
      </svg>
    );
  }
}


function mapStateToProps(state) {
  return {
    viewportWidth: state.getIn(['viewport', 'width']),
  };
}

export default connect(mapStateToProps)(TimeTravelTimeline);
