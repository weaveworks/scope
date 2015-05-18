/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');

var Topologies = React.createClass({

	onTopologyClick: function(ev) {
		ev.preventDefault();
		AppActions.clickTopology(ev.currentTarget.getAttribute('rel'));
	},

	renderTopology: function(topology, active) {
		var className = AppStore.isUrlForTopology(topology.url, active) ? "topologies-item topologies-item-active" : "topologies-item",
			topologyId = AppStore.getTopologyForUrl(topology.url),
			title = ['Topology: ' + topology.name,
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
		var activeTopologyId = this.props.active,
			topologies = _.sortBy(this.props.topologies, function(topology) {
				return topology.name;
			});

		return (
			<div className="topologies">
				{topologies.map(function(topology) {
					return this.renderTopology(topology, activeTopologyId);
				}, this)}
			</div>
		);
	}

});

module.exports = Topologies;
