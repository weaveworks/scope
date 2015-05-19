/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var Logo = require('./logo');
var AppStore = require('../stores/app-store');
var Groupings = require('./groupings.js');
var Status = require('./status.js');
var Topologies = require('./topologies.js');
var TopologyStore = require('../stores/topology-store');
var WebapiUtils = require('../utils/web-api-utils');
var AppActions = require('../actions/app-actions');
var Details = require('./details');
var Nodes = require('./nodes');
var RouterUtils = require('../utils/router-utils');


var ESC_KEY_CODE = 27;

function getStateFromStores() {
	return {
		activeTopology: AppStore.getCurrentTopology(),
		connectionState: AppStore.getConnectionState(),
		currentGrouping: AppStore.getCurrentGrouping(),
		selectedNodeId: AppStore.getSelectedNodeId(),
		nodeDetails: AppStore.getNodeDetails(),
		nodes: TopologyStore.getNodes(),
		topologies: AppStore.getTopologies()
	}
}


var App = React.createClass({

	getInitialState: function() {
		return getStateFromStores();
	},

	componentDidMount: function() {
		TopologyStore.on(TopologyStore.CHANGE_EVENT, this.onChange);
		AppStore.on(AppStore.CHANGE_EVENT, this.onChange);
		window.addEventListener('keyup', this.onKeyPress);

		RouterUtils.getRouter().start({hashbang: true});
		WebapiUtils.getTopologies();
	},

	onChange: function() {
		this.setState(getStateFromStores());
	},

	onKeyPress: function(ev) {
		if (ev.keyCode === ESC_KEY_CODE) {
			AppActions.hitEsc();
		}
	},

	render: function() {
		var showingDetails = this.state.selectedNodeId;

		return (
			<div>
				{showingDetails && <Details nodes={this.state.nodes}
					nodeId={this.state.selectedNodeId}
					details={this.state.nodeDetails}
					topology={this.state.activeTopology} /> }

				<div className="header">
					<Logo />
					<Topologies topologies={this.state.topologies} active={this.state.activeTopology} />
					<Groupings active={this.state.currentGrouping} />
					<Status connectionState={this.state.connectionState} />
				</div>

				<Nodes nodes={this.state.nodes} />
			</div>
		);
	}

});

module.exports = App;
