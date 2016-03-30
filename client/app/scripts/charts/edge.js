import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import { enterEdge, leaveEdge } from '../actions/app-actions';

export default class Edge extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const { hasSelectedNode, highlightedEdgeIds, id, layoutPrecision,
      path, selectedNodeId, source, target } = this.props;

    const classNames = ['edge'];
    if (highlightedEdgeIds.has(id)) {
      classNames.push('highlighted');
    }
    if (hasSelectedNode
      && source !== selectedNodeId
      && target !== selectedNodeId) {
      classNames.push('blurred');
    }
    if (hasSelectedNode && layoutPrecision === 0
      && (source === selectedNodeId || target === selectedNodeId)) {
      classNames.push('focused');
    }
    const classes = classNames.join(' ');

    return (
      <g className={classes} onMouseEnter={this.handleMouseEnter}
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
