import React from 'react';


export default class Tooltip extends React.Component {
  render() {
    return (
      <span className="tooltip-wrapper">
        <div className="tooltip">{this.props.tip}</div>
        {this.props.children}
      </span>
    );
  }
}
