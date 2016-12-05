import React from 'react';


export default class DelayedShow extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      show: false
    };
  }

  componentWillMount() {
    if (this.props.show) {
      this.scheduleShow();
    }
  }

  componentWillUnmount() {
    this.cancelShow();
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.show === this.props.show) {
      return;
    }

    if (nextProps.show) {
      this.scheduleShow();
    } else {
      this.cancelShow();
      this.setState({ show: false });
    }
  }

  scheduleShow() {
    this.showTimeout = setTimeout(() => this.setState({ show: true }), this.props.delay);
  }

  cancelShow() {
    clearTimeout(this.showTimeout);
  }

  render() {
    const { children } = this.props;
    const { show } = this.state;
    const style = {
      opacity: show ? 1 : 0,
      transition: 'opacity 0.5s ease-in-out',
    };
    return (
      <div style={style}>
        {children}
      </div>
    );
  }
}


DelayedShow.defaultProps = {
  delay: 1000
};
