/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');
var NavItemTopology = require('./nav-item-topology.js');

var Topologies = React.createClass({

	onTopologyClick: function(ev) {
		ev.preventDefault();
		AppActions.clickTopology(ev.currentTarget.getAttribute('rel'));
	},

	renderTopology: function(topology, active) {
		var className = AppStore.isUrlForTopology(topology.url, active) ? "nav-preview nav-active" : "nav-preview",
			topologyId = AppStore.getTopologyForUrl(topology.url),
			title = ['Topology: ' + topology.name,
			'Nodes: ' + topology.stats.node_count,
			'Connections: ' + topology.stats.node_count].join('\n');

		return (
			<div className={className} key={topologyId} rel={topologyId} onClick={this.onTopologyClick}>
				<div title={title}>
					<div className="nav-topology-frame">
						<span className="nav-topology-nodes">{topology.stats.node_count}</span>
						<span className="nav-topology-divider" />
						<span className="nav-topology-edges">{topology.stats.edge_count}</span>
					</div>
					<div className="nav-label">
						{topology.name}
					</div>
				</div>
			</div>
		);
	},

	render: function() {
		var activeTopologyId = this.props.active,
			topologies = _.sortBy(this.props.topologies, function(topology) {
				return topology.name;
			});

		return (
			<div className="navbar-topology">
				{topologies.map(function(topology) {
					return this.renderTopology(topology, activeTopologyId);
				}, this)}
			</div>
		);
	}

});

module.exports = Topologies;
