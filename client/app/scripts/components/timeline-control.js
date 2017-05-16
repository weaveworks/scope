import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import { changeTopologyTimestamp } from '../actions/app-actions';


class TimelineControl extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = { value: moment().toISOString() };
    this.handleChange = this.handleChange.bind(this);
    this.queryTopology = debounce(this.queryTopology.bind(this), 2000);
  }

  queryTopology() {
    console.log('QUERY TOPOLOGY AT: ', this.state.value);
    this.props.changeTopologyTimestamp(this.state.value);
  }

  handleChange(ev) {
    const value = ev.target.value;
    this.setState({ value });
    this.queryTopology();
  }

  render() {
    return (
      <div className="timeline-control">
        <input type="datetime" onChange={this.handleChange} value={this.state.value} />
      </div>
    );
  }
}


export default connect(null, { changeTopologyTimestamp })(TimelineControl);
