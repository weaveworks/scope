import React from 'react';
import _ from 'lodash';
import { connect } from 'react-redux';

import NodesChart from '../charts/nodes-chart';
import NodesError from '../charts/nodes-error';
import { findTopologyById, isTopologyEmpty } from '../utils/topology-utils';

const navbarHeight = 160;
const marginTop = 0;
const detailsWidth = 450;


const LOADING_TEMPLATES = [
  'Just loading the THINGS... any second now...',
  "Loading the THINGS. They'll be here in a jiffy...",
  'Crunching the THINGS',
  'Deleting all the THINGS',
  'rm -rf *THINGS*',
  'Waiting for all the THINGS',
  'Containing the THINGS',
  'Processing the THINGS',
];


function renderTemplate(nodeType, template) {
  return template.replace('THINGS', nodeType);
}


/**
 * dynamic coords precision based on topology size
 */
function getLayoutPrecision(nodesCount) {
  let precision;
  if (nodesCount >= 50) {
    precision = 0;
  } else if (nodesCount > 20) {
    precision = 1;
  } else if (nodesCount > 10) {
    precision = 2;
  } else {
    precision = 3;
  }

  return precision;
}

function getNodeType(topology, topologies) {
  if (!topology || topologies.size === 0) {
    return '';
  }
  let name = topology.get('name');
  if (topology.get('parentId')) {
    const parentTopology = findTopologyById(topologies, topology.get('parentId'));
    name = parentTopology.get('name');
  }
  return name.toLowerCase();
}

class Nodes extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleResize = this.handleResize.bind(this);

    const [topologiesLoadingTemplate, nodeLoadingTemplate] = _.sampleSize(LOADING_TEMPLATES, 2);
    this.state = {
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop,
      topologiesLoadingTemplate,
      nodeLoadingTemplate,
    };
  }

  componentDidMount() {
    window.addEventListener('resize', this.handleResize);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
  }

  renderEmptyTopologyError(show) {
    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!show}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        <ul>
          <li>We haven't received any reports from probes recently.
           Are the probes properly configured?</li>
          <li>There are nodes, but they're currently hidden. Check the view options
           in the bottom-left if they allow for showing hidden nodes.</li>
          <li>Containers view only: you're not running Docker,
           or you don't have any containers.</li>
        </ul>
      </NodesError>
    );
  }

  renderLoading(message, show) {
    return (
      <NodesError mainClassName="nodes-chart-loading" faIconClass="fa-circle-thin" hidden={!show}>
        <div className="heading">{message}</div>
      </NodesError>
    );
  }

  render() {
    const { nodes, selectedNodeId, topologyEmpty, topologiesLoaded, nodesLoaded, topologies,
      topology } = this.props;
    const layoutPrecision = getLayoutPrecision(nodes.size);
    const hasSelectedNode = selectedNodeId && nodes.has(selectedNodeId);
    const topologyLoadingMessage = renderTemplate('toplogies',
      this.state.topologiesLoadingTemplate);
    const nodeLoadingMessage = renderTemplate(getNodeType(topology, topologies),
      this.state.nodeLoadingTemplate);

    return (
      <div className="nodes-wrapper">
        {this.renderLoading(topologyLoadingMessage, !topologiesLoaded)}
        {this.renderLoading(nodeLoadingMessage, topologiesLoaded && !nodesLoaded)}
        {this.renderEmptyTopologyError(topologiesLoaded && nodesLoaded && topologyEmpty)}
        <NodesChart {...this.state}
          detailsWidth={detailsWidth}
          layoutPrecision={layoutPrecision}
          hasSelectedNode={hasSelectedNode}
        />
      </div>
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

function mapStateToProps(state) {
  return {
    nodes: state.get('nodes'),
    nodesLoaded: state.get('nodesLoaded'),
    selectedNodeId: state.get('selectedNodeId'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topologyEmpty: isTopologyEmpty(state),
    topology: state.get('currentTopology'),
  };
}

export default connect(
  mapStateToProps
)(Nodes);
