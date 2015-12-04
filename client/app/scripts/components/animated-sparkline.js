// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import d3 from 'd3';
import { OrderedMap } from 'immutable';

import Sparkline from './sparkline';

const makeOrderedMap = OrderedMap;
const parseDate = d3.time.format.iso.parse;

export default class AnimatedSparkline extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.tickTimer = null;
    this.state = {
      buffer: makeOrderedMap(),
      first: null,
      last: null
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
    buffer = buffer.merge(nextSamples);
    const state = {};

    // set first/last marker of sliding window
    if (buffer.size > 0) {
      const bufferKeys = buffer.keySeq();
      if (this.state.first === null) {
        state.first = bufferKeys.first();
      }
      if (this.state.last === null) {
        state.last = bufferKeys.last();
      }
    }

    // remove old values from buffer
    const first = this.state.first ? this.state.first : state.first;
    state.buffer = buffer.skipWhile((v, d) => d < first);

    return state;
  }

  tick() {
    if (this.state.last < this.state.buffer.keySeq().last()) {
      const dates = this.state.buffer.keySeq();
      let firstIndex = dates.indexOf(this.state.first);
      if (firstIndex > -1 && firstIndex < dates.size - 1) {
        firstIndex++;
      } else {
        firstIndex = 0;
      }
      const first = dates.get(firstIndex);

      let lastIndex = dates.indexOf(this.state.last);
      if (lastIndex > -1) {
        lastIndex++;
      } else {
        lastIndex = dates.length - 1;
      }
      const last = dates.get(lastIndex);

      this.tickTimer = setTimeout(() => {
        this.tickTimer = null;
        this.setState({first, last});
      }, 900);
    }
  }

  getGraphData() {
    let first = this.state.first;
    if (this.props.first && this.props.first < this.state.first) {
      // first prop date is way before buffer, keeping it
      first = this.props.first;
    }
    let last = this.state.last;
    if (this.props.last && this.props.last > this.state.buffer.keySeq().last()) {
      // prop last is after buffer values, need to shift dates
      const skip = parseDate(this.props.last) - parseDate(this.state.buffer.keySeq().last());
      last -= skip;
      first -= skip;
    }
    const dateFilter = d => d.date >= first && d.date <= last;
    const data = this.state.buffer.map((v, k) => {
      return {value: v, date: k};
    }).toIndexedSeq().toJS().filter(dateFilter);

    return {first, last, data};
  }

  render() {
    const {data, first, last} = this.getGraphData();

    return (
      <Sparkline data={data} first={first} last={last} min={this.props.min} />
    );
  }

}

AnimatedSparkline.propTypes = {
  data: React.PropTypes.array.isRequired
};
