import React from 'react';

import NodesChart from '../charts/nodes-chart';

const navbarHeight = 160;
const marginTop = 0;

export default class Nodes extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleResize = this.handleResize.bind(this);

    this.state = {
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop
    };
  }

  componentDidMount() {
    window.addEventListener('resize', this.handleResize);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
  }

  render() {
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
  }

  handleResize() {
    this.setDimensions();
  }

  setDimensions() {
    const width = window.innerWidth;
    const height = window.innerHeight - navbarHeight - marginTop;

    this.setState({height, width});
  }
}
