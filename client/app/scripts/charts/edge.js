import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { enterEdge, leaveEdge } from '../actions/app-actions';

class Edge extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const { id, path, highlighted, blurred, focused } = this.props;
    const className = classNames('edge', {highlighted, blurred, focused});

    return (
      <g
        className={className} onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave} id={id}>
        <path d={path} className="shadow" />
        <path d={path} className="link" />
      </g>
    );
  }

  handleMouseEnter(ev) {
    this.props.enterEdge(ev.currentTarget.id);
  }

  handleMouseLeave(ev) {
    this.props.leaveEdge(ev.currentTarget.id);
  }
}

export default connect(
  null,
  { enterEdge, leaveEdge }
)(Edge);
