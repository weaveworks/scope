import React from 'react';
import { connect } from 'react-redux';
import classnames from 'classnames';

import { searchNodeMatchesSelector } from '../selectors/search';
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
    const isActive = subTopology === this.props.currentTopology;
    const topologyId = subTopology.get('id');
    const searchMatches = this.props.searchNodeMatches.get(subTopology.get('id'));
    const searchMatchCount = searchMatches ? searchMatches.size : 0;
    const title = basicTopologyInfo(subTopology, searchMatchCount);
    const className = classnames('topologies-sub-item', {
      'topologies-sub-item-active': isActive,
      'topologies-sub-item-matched': searchMatchCount
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
    const searchMatches = this.props.searchNodeMatches.get(topology.get('id'));
    const searchMatchCount = searchMatches ? searchMatches.size : 0;
    const className = classnames('topologies-item-main', {
      'topologies-item-main-active': isActive,
      'topologies-item-main-matched': searchMatchCount
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
    // searchNodeMatches: state.get('searchNodeMatches'),
    searchNodeMatches: searchNodeMatchesSelector(state),
    currentTopology: state.get('currentTopology')
  };
}

export default connect(
  mapStateToProps,
  { clickTopology }
)(Topologies);
