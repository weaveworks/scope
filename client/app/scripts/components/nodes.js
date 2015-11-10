const d3 = require('d3');
const React = require('react');

const NodesChart = require('../charts/nodes-chart');

const navbarHeight = 160;
const marginTop = 0;

const Nodes = React.createClass({

  getInitialState: function() {
    return {
      nodeScale: d3.scale.linear(),
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop
    };
  },

  componentDidMount: function() {
    this.setDimensions();
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
        nodeScale={this.state.nodeScale}
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
    const expanse = Math.min(height, width);
    const nodeSize = expanse / 3; // single node should fill a third of the screen
    const maxNodeSize = expanse / 10;
    const normalizedNodeSize = Math.min(nodeSize / Math.sqrt(this.props.nodes.size), maxNodeSize);
    const nodeScale = this.state.nodeScale.range([0, normalizedNodeSize]);

    this.setState({height, width, nodeScale});
  }

});

module.exports = Nodes;
