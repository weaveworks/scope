const React = require('react');
const _ = require('lodash');

const NodesChart = require('../charts/nodes-chart');
const NodeDetails = require('./node-details');

const marginBottom = 64;
const marginTop = 64;
const marginLeft = 36;
const marginRight = 36;

const Explorer = React.createClass({

  getInitialState: function() {
    return {
      layout: 'solar',
      width: window.innerWidth - marginLeft - marginRight,
      height: window.innerHeight - marginBottom - marginTop
    };
  },

  componentDidMount: function() {
    window.addEventListener('resize', this.handleResize);
  },

  componentWillUnmount: function() {
    window.removeEventListener('resize', this.handleResize);
  },

  getSubTopology: function(topology) {
    const subTopology = {};
    const nodeSet = [];

    _.each(this.props.expandedNodes, function(nodeId) {
      if (topology[nodeId]) {
        subTopology[nodeId] = topology[nodeId];
        nodeSet = _.union(subTopology[nodeId].adjacency, nodeSet);
        _.each(subTopology[nodeId].adjacency, function(adjacentId) {
          const node = _.assign({}, topology[adjacentId]);

          subTopology[adjacentId] = node;
        });
      }
    });

    // weed out edges
    _.each(subTopology, function(node) {
      node.adjacency = _.intersection(node.adjacency, nodeSet);
    });

    return subTopology;
  },

  render: function() {
    const subTopology = this.getSubTopology(this.props.nodes);

    return (
      <div id="explorer">
        <NodeDetails details={this.props.details} />
        <div className="graph">
          <NodesChart
            layout={this.state.layout}
            nodes={subTopology}
            highlightedNodes={this.props.expandedNodes}
            width={this.state.width}
            height={this.state.height}
            context="explorer"
          />
        </div>
      </div>
    );
  },

  setDimensions: function() {
    this.setState({
      height: window.innerHeight - marginBottom - marginTop,
      width: window.innerWidth - marginLeft - marginRight
    });
  },

  handleResize: function() {
    this.setDimensions();
  }

});

module.exports = Explorer;
