/** @jsx React.DOM */

var React = require('react');

var NodesChart = require('../charts/nodes-chart');
var AppActions = require('../actions/app-actions');

var navbarHeight = 160;
var marginTop = 0;
var marginLeft = 0;

var Nodes = React.createClass({

  getInitialState: function() {
    return {
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop
    };
  },

  onNodeClick: function(ev) {
    AppActions.clickNode(ev.currentTarget.id);
  },

  componentDidMount: function() {
    window.addEventListener('resize', this.handleResize);
  },

  componentWillUnmount: function() {
    window.removeEventListener('resize', this.handleResize);
  },

  setDimensions: function() {
    this.setState({
      height: window.innerHeight - navbarHeight - marginTop,
      width: window.innerWidth
    });
  },

  handleResize: function() {
    this.setDimensions();
  },

  render: function() {
    return (
      <div id="nodes">
        <NodesChart
          onNodeClick={this.onNodeClick}
          nodes={this.props.nodes}
          width={this.state.width}
          height={this.state.height}
          context="view"
        />
      </div>
    );
  }

});

module.exports = Nodes;