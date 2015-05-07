/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');
var mui = require('material-ui');
var Paper = mui.Paper;

var NodeDetails = require('./node-details');
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
				<Paper>
					<NodeDetails details={this.props.details} />
				</Paper>
			</div>
		);
	}

});

module.exports = Details;