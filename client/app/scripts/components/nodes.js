const React = require('react');

const NodesChart = require('../charts/nodes-chart');
const AppActions = require('../actions/app-actions');

const navbarHeight = 160;
const marginTop = 0;

const Nodes = React.createClass({

  getInitialState: function() {
    return {
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop
    };
  },

  componentDidMount: function() {
    window.addEventListener('resize', this.handleResize);
  },

  componentWillUnmount: function() {
    window.removeEventListener('resize', this.handleResize);
  },

  onNodeClick: function(ev) {
    AppActions.clickNode(ev.currentTarget.id);
  },

  render: function() {
    return (
      <div id="nodes">
        <NodesChart
          highlightedEdgeIds={this.props.highlightedEdgeIds}
          highlightedNodeIds={this.props.highlightedNodeIds}
          nodes={this.props.nodes}
          onNodeClick={this.onNodeClick}
          width={this.state.width}
          height={this.state.height}
          topologyId={this.props.topologyId}
          context="view"
        />
      </div>
    );
  },

  handleResize: function() {
    this.setDimensions();
  },

  setDimensions: function() {
    this.setState({
      height: window.innerHeight - navbarHeight - marginTop,
      width: window.innerWidth
    });
  }

});

module.exports = Nodes;
