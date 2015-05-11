/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var NodeDetailsTable = React.createClass({

	render: function() {
		return (
			<div className="node-details-table">
				<h4 className="node-details-table-title">
					{this.props.title}
				</h4>

				{this.props.rows.map(function(row) {
					return (
						<div className="node-details-table-row">
							<div className="node-details-table-row-key">{row.key}</div>
							<div className="node-details-table-row-value-major">{row.value_major}</div>
							<div className="node-details-table-row-value-minor">{row.value_minor}</div>
						</div>
					);
				})}
        	</div>
		);
	}

});

module.exports = NodeDetailsTable;