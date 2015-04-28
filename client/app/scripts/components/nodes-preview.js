/** @jsx React.DOM */

var React = require('react');

var NodesChart = require('../charts/nodes-chart');

var NodesPreview = React.createClass({

	render: function() {
		return (
			<div id="nodes-preview">
				<div className="graph">
					<NodesChart
						layout="circle"
						nodes={this.props.nodes}
						width={62}
						height={62}
						context="preview"
					/>
				</div>
			</div>
		);
	}

});

module.exports = NodesPreview;