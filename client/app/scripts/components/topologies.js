import React from 'react';
import _ from 'lodash';

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
    const isActive = subTopology.name === this.props.currentTopology.name;
    const topologyId = subTopology.id;
    const title = this.renderTitle(subTopology);
    const className = isActive ? 'topologies-sub-item topologies-sub-item-active' : 'topologies-sub-item';

    return (
      <div className={className} title={title} key={topologyId} rel={topologyId}
        onClick={this.onTopologyClick}>
        <div className="topologies-sub-item-label">
          {subTopology.name}
        </div>
      </div>
    );
  }

  renderTitle(topology) {
    return ['Nodes: ' + topology.stats.node_count,
      'Connections: ' + topology.stats.node_count].join('\n');
  }

  renderTopology(topology) {
    const isActive = topology.name === this.props.currentTopology.name;
    const className = isActive ? 'topologies-item-main topologies-item-main-active' : 'topologies-item-main';
    const topologyId = topology.id;
    const title = this.renderTitle(topology);

    return (
      <div className="topologies-item" key={topologyId}>
        <div className={className} title={title} rel={topologyId} onClick={this.onTopologyClick}>
          <div className="topologies-item-label">
            {topology.name}
          </div>
        </div>
        <div className="topologies-sub">
          {topology.sub_topologies && topology.sub_topologies.map(this.renderSubTopology)}
        </div>
      </div>
    );
  }

  render() {
    const topologies = _.sortBy(this.props.topologies, function(topology) {
      return topology.rank;
    });

    return (
      <div className="topologies">
        {this.props.currentTopology && topologies.map(function(topology) {
          return this.renderTopology(topology);
        }, this)}
      </div>
    );
  }
}
