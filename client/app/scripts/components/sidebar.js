import React from 'react';

export default class Sidebar extends React.Component {
  render() {
    return (
      <div className="sidebar">
        {this.props.children}
      </div>
    );
  }
}
