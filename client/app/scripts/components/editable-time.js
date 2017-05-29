import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';

import { ENTER_KEY_CODE } from '../constants/key-codes';
import { changeTopologyTimestamp } from '../actions/app-actions';


class EditableTime extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = {
      value: '',
      editing: false,
      timeOffset: 0,
    };

    this.handleFocus = this.handleFocus.bind(this);
    this.handleBlur = this.handleBlur.bind(this);
    this.handleKeyUp = this.handleKeyUp.bind(this);
    this.handleChange = this.handleChange.bind(this);
    this.saveInputRef = this.saveInputRef.bind(this);
  }

  componentDidMount() {
    this.timer = setInterval(() => {
      if (!this.state.editing) {
        this.setState(this.getFreshState());
      }
    }, 100);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  handleFocus() {
    this.setState({ editing: true });
  }

  handleBlur() {
    this.setState({ editing: false });
  }

  handleChange(ev) {
    this.setState({ value: ev.target.value });
  }

  handleKeyUp(ev) {
    if (ev.keyCode === ENTER_KEY_CODE) {
      const { value } = this.state;
      const timeOffset = moment.duration(moment().diff(value));
      this.props.changeTopologyTimestamp(value);

      this.setState({ timeOffset });
      this.inputRef.blur();

      console.log('QUERY TOPOLOGY AT: ', value, timeOffset);
    }
  }

  saveInputRef(ref) {
    this.inputRef = ref;
  }

  getFreshState() {
    const { timeOffset } = this.state;
    const time = moment().utc().subtract(timeOffset);
    return { value: time.toISOString() };
  }

  render() {
    return (
      <span className="running-time">
        <input
          type="text"
          onFocus={this.handleFocus}
          onBlur={this.handleBlur}
          onKeyUp={this.handleKeyUp}
          onChange={this.handleChange}
          value={this.state.value}
          ref={this.saveInputRef}
        />
      </span>
    );
  }
}

export default connect(null, { changeTopologyTimestamp })(EditableTime);
