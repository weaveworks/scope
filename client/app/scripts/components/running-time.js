import React from 'react';
import moment from 'moment';


export default class RunningTime extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = this.getFreshState();
  }

  componentDidMount() {
    this.timer = setInterval(() => {
      if (!this.props.paused) {
        this.setState(this.getFreshState());
      }
    }, 500);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  getFreshState() {
    const timestamp = moment().utc().subtract(this.props.offsetMilliseconds);
    return { humanizedTimestamp: timestamp.format('MMMM Do YYYY, h:mm:ss a') };
  }

  render() {
    return (
      <span className="running-time">{this.state.humanizedTimestamp}</span>
    );
  }
}
