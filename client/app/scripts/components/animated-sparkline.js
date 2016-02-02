// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import d3 from 'd3';
import { OrderedMap } from 'immutable';

import Sparkline from './sparkline';

const makeOrderedMap = OrderedMap;
const parseDate = d3.time.format.iso.parse;
const sortDate = (v, d) => d;

export default class AnimatedSparkline extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.tickTimer = null;
    this.state = {
      buffer: makeOrderedMap(),
      movingFirst: null,
      movingLast: null
    };
  }

  componentWillMount() {
    this.setState(this.updateBuffer(this.props));
  }

  componentWillUnmount() {
    clearTimeout(this.tickTimer);
  }

  componentWillReceiveProps(nextProps) {
    this.setState(this.updateBuffer(nextProps));
  }

  componentDidUpdate() {
    // move sliding window one tick
    if (!this.tickTimer && this.state.buffer.size > 0) {
      this.tick();
    }
  }

  updateBuffer(props) {
    // merge new samples into buffer
    let buffer = this.state.buffer;
    const nextSamples = makeOrderedMap(props.data.map(d => [d.date, d.value]));
    buffer = buffer.merge(nextSamples).sortBy(sortDate);
    const state = {};

    // set first/last marker of sliding window
    if (buffer.size > 0) {
      const bufferKeys = buffer.keySeq();
      if (this.state.movingFirst === null) {
        state.movingFirst = bufferKeys.first();
      }
      if (this.state.movingLast === null) {
        state.movingLast = bufferKeys.last();
      }
    }

    // remove old values from buffer
    const movingFirst = this.state.movingFirst ? this.state.movingFirst : state.movingFirst;
    state.buffer = buffer.filter((v, d) => d >= movingFirst);

    return state;
  }

  tick() {
    const { buffer } = this.state;
    let { movingFirst, movingLast } = this.state;
    const bufferKeys = buffer.keySeq();

    if (movingLast < bufferKeys.last()) {
      let firstIndex = bufferKeys.indexOf(movingFirst);
      if (firstIndex > -1 && firstIndex < bufferKeys.size - 1) {
        firstIndex++;
      } else {
        firstIndex = 0;
      }
      movingFirst = bufferKeys.get(firstIndex);

      let lastIndex = bufferKeys.indexOf(movingLast);
      if (lastIndex > -1) {
        lastIndex++;
      } else {
        lastIndex = bufferKeys.length - 1;
      }
      movingLast = bufferKeys.get(lastIndex);

      this.tickTimer = setTimeout(() => {
        this.tickTimer = null;
        this.setState({movingFirst, movingLast});
      }, 900);
    }
  }

  getGraphData() {
    const firstDate = parseDate(this.props.first);
    const lastDate = parseDate(this.props.last);
    const { buffer } = this.state;
    let movingFirstDate = parseDate(this.state.movingFirst);
    let movingLastDate = parseDate(this.state.movingLast);
    const lastBufferDate = parseDate(buffer.keySeq().last());

    if (firstDate && movingFirstDate && firstDate < movingFirstDate) {
      // first prop date is way before buffer, keeping it
      movingFirstDate = firstDate;
    }
    if (lastDate && lastBufferDate && lastDate > lastBufferDate) {
      // prop last is after buffer values, need to shift dates
      const skip = lastDate - lastBufferDate;
      movingLastDate -= skip;
      movingFirstDate -= skip;
    }
    const dateFilter = d => d.date >= movingFirstDate && d.date <= movingLastDate;
    const data = this.state.buffer
      .map((v, k) => ({value: v, date: +parseDate(k)}))
      .toIndexedSeq()
      .toJS()
      .filter(dateFilter);

    return {movingFirstDate, movingLastDate, data};
  }

  render() {
    const {data, movingFirstDate, movingLastDate} = this.getGraphData();

    return (
      <Sparkline data={data} first={movingFirstDate} last={movingLastDate} min={this.props.min} />
    );
  }

}

AnimatedSparkline.propTypes = {
  data: React.PropTypes.array.isRequired
};
