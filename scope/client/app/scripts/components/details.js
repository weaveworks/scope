/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var DetailViews = require('./detail-views');
var Explorer = require('./explorer');
var WebapiUtils = require('../utils/web-api-utils');

var Details = React.createClass({

	getNodeDetails: function(nodes, nodeId) {
		var node = nodes[nodeId];

		WebapiUtils.getNodeDetails(this.props.topology, nodeId);
	},

	componentDidMount: function() {
		this.getNodeDetails(this.props.nodes, this.props.explorerExpandedNodes[0]);
	},

	render: function() {
		return (
			<div id="details">
				<DetailViews active={this.props.view} details={this.props.details} />
				<Explorer nodes={this.props.nodes} details={this.props.details}
					expandedNodes={this.props.explorerExpandedNodes} />
			</div>
		);
	}

});

module.exports = Details;