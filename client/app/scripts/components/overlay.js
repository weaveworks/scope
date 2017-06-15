import React from 'react';
import classNames from 'classnames';


export default class Overlay extends React.Component {
  render() {
    const className = classNames('overlay', { faded: this.props.faded });

    return <div className={className} />;
  }
}
