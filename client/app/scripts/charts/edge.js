import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { enterEdge, leaveEdge } from '../actions/app-actions';
import { isContrastMode } from '../utils/contrast-utils';
import { NODE_BASE_SIZE } from '../constants/styles';

class Edge extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const { id, path, highlighted, blurred, focused, scale } = this.props;
    const className = classNames('edge', { highlighted, blurred, focused });
    const thickness = scale * (isContrastMode() ? 0.015 : 0.0075) * NODE_BASE_SIZE;

    // Draws the edge so that its thickness reflects the zoom scale.
    // Edge shadow is always made 10x thicker than the edge itself.
    return (
      <g
        id={id} className={className}
        onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave}>
        <path className="shadow" d={path} style={{ strokeWidth: 10 * thickness }} />
        <path className="link" d={path} style={{ strokeWidth: thickness }} />
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
