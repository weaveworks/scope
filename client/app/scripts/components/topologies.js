import React from 'react';
import { connect } from 'react-redux';
import classnames from 'classnames';

import { searchMatchCountByTopologySelector } from '../selectors/search';
import { isResourceViewModeSelector } from '../selectors/topology';
import { clickTopology } from '../actions/app-actions';


function basicTopologyInfo(topology, searchMatchCount) {
  const info = [
    `Nodes: ${topology.getIn(['stats', 'node_count'])}`,
    `Connections: ${topology.getIn(['stats', 'edge_count'])}`
  ];
  if (searchMatchCount) {
    info.push(`Search Matches: ${searchMatchCount}`);
  }
  return info.join('\n');
}

class Topologies extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onTopologyClick = this.onTopologyClick.bind(this);
  }

  onTopologyClick(ev) {
    ev.preventDefault();
    this.props.clickTopology(ev.currentTarget.getAttribute('rel'));
  }

  renderSubTopology(subTopology) {
    const topologyId = subTopology.get('id');
    const isActive = subTopology === this.props.currentTopology;
    const searchMatchCount = this.props.searchMatchCountByTopology.get(topologyId) || 0;
    const title = basicTopologyInfo(subTopology, searchMatchCount);
    const className = classnames('topologies-sub-item', {
      // Don't show matches in the resource view as searching is not supported there yet.
      'topologies-sub-item-matched': !this.props.isResourceViewMode && searchMatchCount,
      'topologies-sub-item-active': isActive,
    });

    return (
      <div
        className={className} title={title} key={topologyId} rel={topologyId}
        onClick={this.onTopologyClick}>
        <div className="topologies-sub-item-label">
          {subTopology.get('name')}
        </div>
      </div>
    );
  }

  renderTopology(topology) {
    const isActive = topology === this.props.currentTopology;
    const searchMatchCount = this.props.searchMatchCountByTopology.get(topology.get('id')) || 0;
    const className = classnames('topologies-item-main', {
      // Don't show matches in the resource view as searching is not supported there yet.
      'topologies-item-main-matched': !this.props.isResourceViewMode && searchMatchCount,
      'topologies-item-main-active': isActive,
    });
    const topologyId = topology.get('id');
    const title = basicTopologyInfo(topology, searchMatchCount);

    return (
      <div className="topologies-item" key={topologyId}>
        <div className={className} title={title} rel={topologyId} onClick={this.onTopologyClick}>
          <div className="topologies-item-label">
            {topology.get('name')}
          </div>
        </div>
        <div className="topologies-sub">
          {topology.has('sub_topologies')
            && topology.get('sub_topologies').map(subTop => this.renderSubTopology(subTop))}
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

function mapStateToProps(state) {
  return {
    topologies: state.get('topologies'),
    currentTopology: state.get('currentTopology'),
    searchMatchCountByTopology: searchMatchCountByTopologySelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { clickTopology }
)(Topologies);
