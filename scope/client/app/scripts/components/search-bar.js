/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var Stats = require('./stats.js');
var TopologyActions = require('../actions/topology-actions');

var SearchBar = React.createClass({

	getDefaultProps: function() {
		return {
			initialFilterText: ''
		};
	},

	getInitialState: function() {
		return {
			filterAdjacent: false,
			filterText: this.props.filterText
		};
	},

	handleChange: function() {
		TopologyActions.inputFilterText(this.state.filterText);
	},

	onVolatileChange: function(event) {
		this.setState({
			filterText: event.target.value
		});

		this.scheduleChange();
	},

	scheduleChange: _.debounce(function() {
		this.handleChange();
	}, 300),

	onFilterAjacent: function(event) {
		this.setState({
			filterAdjacent: event.target.checked
		});
		TopologyActions.checkFilterAdjacent(event.target.checked);
	},

	componentWillMount: function() {
		this.setState({
			filterText: this.props.filterText
		});
	},

	componentDidMount: function() {
		this.refs.filterTextInput.getDOMNode().focus();
	},

	render: function() {
		return (
			<div className="row" id="search-bar">
				<div className="form-group col-md-9">
					<div className="form-control-wrapper">
						<input className="form-control input-lg" type="text"
							id="filterTextInput"
							placeholder="Filter for nodes"
							value={this.state.filterText}
							ref="filterTextInput"
							onChange={this.onVolatileChange} />
						<span className="material-input"></span>
					</div>
				</div>
				<div className="form-group col-md-3">
					<div className="checkbox">
                        <label>
                            <input type="checkbox" 
                            	ref="filterAdjacent"
                            	value={this.state.filterAdjacent}
                            	onChange={this.onFilterAjacent} />
                            <span className="check"></span>
                            Include connected
                        </label>
                    </div>
				</div>
			</div>
		);
	}

});

module.exports = SearchBar;