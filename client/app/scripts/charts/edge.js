import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';
import classNames from 'classnames';

import { enterEdge, leaveEdge } from '../actions/app-actions';

export default class Edge extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const { id, path, highlighted, blurred, focused } = this.props;
    const className = classNames('edge', {highlighted, blurred, focused});

    return (
      <g className={className} onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave} id={id}>
        <path d={path} className="shadow" />
        <path d={path} className="link" />
      </g>
    );
  }

  handleMouseEnter(ev) {
    enterEdge(ev.currentTarget.id);
  }

  handleMouseLeave(ev) {
    leaveEdge(ev.currentTarget.id);
  }
}

reactMixin.onClass(Edge, PureRenderMixin);
