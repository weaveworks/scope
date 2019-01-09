import React from 'react';
import { isoParse as parseDate } from 'd3-time-format';
import { OrderedMap } from 'immutable';

const makeOrderedMap = OrderedMap;
const sortDate = (v, d) => d;
const DEFAULT_TICK_INTERVAL = 1000; // DEFAULT_TICK_INTERVAL + renderTime < 1000ms
const WINDOW_LENGTH = 60;

/**
 * Higher-order component that buffers a metrics series and feeds a sliding
 * window of the series to the wrapped component.
 *
 * Initial samples `[t0, t1, t2, ...]` will be passed as is. When new data
 * `[t2, t3, t4, ...]` comes in, it will be merged into the buffer:
 * `[t0, t1, t2, t3, t4, ...]`. On next `tick()` the window shifts and
 * `[t1, t2, t3, ...]` will be fed to the wrapped component.
 *  The window slides between the dates provided by the first date of the buffer
 *  and `this.props.last` so that the following invariant is true:
 * `this.state.movingFirst <= this.props.first < this.state.movingLast <= this.props.last`.
 *
 * Samples have to be of type `[{date: String, value: Number}, ...]`.
 * This component also keeps a historic max of all samples it sees over time.
 */
export default ComposedComponent => (class extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.tickTimer = null;
    this.state = {
      buffer: makeOrderedMap(),
      max: 0,
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
    this.tick();
  }

  componentDidMount() {
    this.tick();
  }

  updateBuffer(props) {
    // merge new samples into buffer
    let { buffer } = this.state;
    const nextSamples = makeOrderedMap(props.samples.map(d => [d.date, d.value]));
    // need to sort again after merge, some new data may have different times for old values
    buffer = buffer.merge(nextSamples).sortBy(sortDate);
    const state = {};

    // remove old values from buffer
    if (this.state.movingFirst !== null) {
      buffer = buffer.filter((v, d) => d > this.state.movingFirst);
    }
    state.buffer = buffer;

    // set historic max
    state.max = Math.max(buffer.max(), this.state.max);

    // set first/last marker of sliding window
    if (buffer.size > 1) {
      const bufferKeys = buffer.keySeq();
      const firstPart = bufferKeys.slice(0, Math.floor(buffer.size / 3));

      if (this.state.movingFirst === null) {
        state.movingFirst = firstPart.first();
      }
      if (this.state.movingLast === null) {
        state.movingLast = firstPart.last();
      }
    }

    return state;
  }

  tick() {
    // only tick after setTimeout -> setState -> componentDidUpdate
    if (!this.tickTimer) {
      const { buffer } = this.state;
      let { movingFirst, movingLast } = this.state;
      const bufferKeys = buffer.keySeq();

      // move the sliding window one tick, make sure to keep WINDOW_LENGTH values
      if (buffer.size > 0 && movingLast < bufferKeys.last()) {
        let firstIndex = bufferKeys.indexOf(movingFirst);
        let lastIndex = bufferKeys.indexOf(movingLast);

        // speed up the window if it falls behind
        const step = lastIndex > 0 ? Math.round(buffer.size / lastIndex) : 1;

        // only move first if we have enough values in window
        const windowLength = lastIndex - firstIndex;
        if (firstIndex > -1 && firstIndex < bufferKeys.size - 1 && windowLength >= WINDOW_LENGTH) {
          firstIndex += step + (windowLength - WINDOW_LENGTH);
        } else {
          firstIndex = 0;
        }
        movingFirst = bufferKeys.get(firstIndex);
        if (!movingFirst) {
          movingFirst = bufferKeys.first();
        }

        if (lastIndex > -1) {
          lastIndex += step;
        } else {
          lastIndex = bufferKeys.size - 1;
        }
        movingLast = bufferKeys.get(lastIndex);
        if (!movingLast) {
          movingLast = bufferKeys.last();
        }

        this.tickTimer = setTimeout(() => {
          this.tickTimer = null;
          this.setState({movingFirst, movingLast});
        }, DEFAULT_TICK_INTERVAL);
      }
    }
  }

  render() {
    const { buffer, max } = this.state;
    const movingFirstDate = parseDate(this.state.movingFirst);
    const movingLastDate = parseDate(this.state.movingLast);

    const dateFilter = d => d.date > movingFirstDate && d.date <= movingLastDate;
    const samples = buffer
      .map((v, k) => ({date: +parseDate(k), value: v}))
      .toIndexedSeq()
      .toJS()
      .filter(dateFilter);

    const lastValue = samples.length > 0 ? samples[samples.length - 1].value : null;
    const slidingWindow = {
      first: movingFirstDate,
      last: movingLastDate,
      max,
      samples,
      value: lastValue
    };

    return <ComposedComponent {...this.props} {...slidingWindow} />;
  }
});
