const React = require('react');

const NodesChart = require('../charts/nodes-chart');

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

  render: function() {
    return (
      <NodesChart
        highlightedEdgeIds={this.props.highlightedEdgeIds}
        highlightedNodeIds={this.props.highlightedNodeIds}
        selectedNodeId={this.props.selectedNodeId}
        nodes={this.props.nodes}
        width={this.state.width}
        height={this.state.height}
        topologyId={this.props.topologyId}
        detailsWidth={this.props.detailsWidth}
        topMargin={this.props.topMargin}
      />
    );
  },

  handleResize: function() {
    this.setDimensions();
  },

  setDimensions: function() {
    const width = window.innerWidth;
    const height = window.innerHeight - navbarHeight - marginTop;

    this.setState({height, width});
  }

});

module.exports = Nodes;
