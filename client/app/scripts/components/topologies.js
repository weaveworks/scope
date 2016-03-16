import React from 'react';

import { clickTopology } from '../actions/app-actions';

export default class Topologies extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onTopologyClick = this.onTopologyClick.bind(this);
    this.renderSubTopology = this.renderSubTopology.bind(this);
  }

  onTopologyClick(ev) {
    ev.preventDefault();
    clickTopology(ev.currentTarget.getAttribute('rel'));
  }

  renderSubTopology(subTopology) {
    const isActive = subTopology === this.props.currentTopology;
    const topologyId = subTopology.get('id');
    const title = this.renderTitle(subTopology);
    const className = isActive
      ? 'topologies-sub-item topologies-sub-item-active' : 'topologies-sub-item';

    return (
      <div className={className} title={title} key={topologyId} rel={topologyId}
        onClick={this.onTopologyClick}>
        <div className="topologies-sub-item-label">
          {subTopology.get('name')}
        </div>
      </div>
    );
  }

  renderTitle(topology) {
    return `Nodes: ${topology.getIn(['stats', 'node_count'])}\n`
      + `Connections: ${topology.getIn(['stats', 'node_count'])}`;
  }

  renderTopology(topology) {
    const isActive = topology === this.props.currentTopology;
    const className = isActive
      ? 'topologies-item-main topologies-item-main-active' : 'topologies-item-main';
    const topologyId = topology.get('id');
    const title = this.renderTitle(topology);

    return (
      <div className="topologies-item" key={topologyId}>
        <div className={className} title={title} rel={topologyId} onClick={this.onTopologyClick}>
          <div className="topologies-item-label">
            {topology.get('name')}
          </div>
        </div>
        <div className="topologies-sub">
          {topology.has('sub_topologies')
            && topology.get('sub_topologies').map(this.renderSubTopology)}
        </div>
      </div>
    );
  }

  render() {
    return (
      <div className="topologies">
        {this.props.currentTopology && this.props.topologies.map(
          topology => this.renderTopology(topology)
        )}
      </div>
    );
  }
}
