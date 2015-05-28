const React = require('react');
const _ = require('lodash');

const AppActions = require('../actions/app-actions');
const AppStore = require('../stores/app-store');

const Topologies = React.createClass({

  onTopologyClick: function(ev) {
    ev.preventDefault();
    AppActions.clickTopology(ev.currentTarget.getAttribute('rel'));
  },

  renderTopology: function(topology) {
    const isActive = topology.name === this.props.currentTopology.name;
    const className = isActive ? 'topologies-item topologies-item-active' : 'topologies-item';
    const topologyId = AppStore.getTopologyIdForUrl(topology.url);
    const title = ['Topology: ' + topology.name,
      'Nodes: ' + topology.stats.node_count,
      'Connections: ' + topology.stats.node_count].join('\n');

    return (
      <div className={className} key={topologyId} rel={topologyId} onClick={this.onTopologyClick}>
        <div title={title}>
          <div className="topologies-item-label">
            {topology.name}
          </div>
        </div>
      </div>
    );
  },

  render: function() {
    const topologies = _.sortBy(this.props.topologies, function(topology) {
        return topology.name;
      });

    return (
      <div className="topologies">
        {this.props.currentTopology && topologies.map(function(topology) {
          return this.renderTopology(topology);
        }, this)}
      </div>
    );
  }

});

module.exports = Topologies;
