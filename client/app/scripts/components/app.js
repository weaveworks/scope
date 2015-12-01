import React from 'react';

import Logo from './logo';
import AppStore from '../stores/app-store';
import Sidebar from './sidebar.js';
import Status from './status.js';
import Topologies from './topologies.js';
import TopologyOptions from './topology-options.js';
import { getApiDetails, getTopologies } from '../utils/web-api-utils';
import { hitEsc } from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import { getRouter } from '../utils/router-utils';

const ESC_KEY_CODE = 27;

function getStateFromStores() {
  return {
    activeTopologyOptions: AppStore.getActiveTopologyOptions(),
    controlError: AppStore.getControlError(),
    controlPending: AppStore.isControlPending(),
    currentTopology: AppStore.getCurrentTopology(),
    currentTopologyId: AppStore.getCurrentTopologyId(),
    currentTopologyOptions: AppStore.getCurrentTopologyOptions(),
    errorUrl: AppStore.getErrorUrl(),
    highlightedEdgeIds: AppStore.getHighlightedEdgeIds(),
    highlightedNodeIds: AppStore.getHighlightedNodeIds(),
    selectedNodeId: AppStore.getSelectedNodeId(),
    nodeDetails: AppStore.getNodeDetails(),
    nodes: AppStore.getNodes(),
    topologies: AppStore.getTopologies(),
    topologiesLoaded: AppStore.isTopologiesLoaded(),
    version: AppStore.getVersion(),
    websocketClosed: AppStore.isWebsocketClosed()
  };
}


export default class App extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);
    this.onKeyPress = this.onKeyPress.bind(this);
    this.state = getStateFromStores();
  }

  componentDidMount() {
    AppStore.addListener(this.onChange);
    window.addEventListener('keyup', this.onKeyPress);

    getRouter().start({hashbang: true});
    if (!AppStore.isRouteSet()) {
      // dont request topologies when already done via router
      getTopologies(AppStore.getActiveTopologyOptions());
    }
    getApiDetails();
  }

  onChange() {
    this.setState(getStateFromStores());
  }

  onKeyPress(ev) {
    if (ev.keyCode === ESC_KEY_CODE) {
      hitEsc();
    }
  }

  render() {
    const showingDetails = this.state.selectedNodeId;
    const versionString = this.state.version ? 'Version ' + this.state.version : '';
    // width of details panel blocking a view
    const detailsWidth = showingDetails ? 450 : 0;
    const topMargin = 100;

    return (
      <div>
        {showingDetails && <Details nodes={this.state.nodes}
          controlError={this.state.controlError}
          controlPending={this.state.controlPending}
          nodeId={this.state.selectedNodeId}
          details={this.state.nodeDetails} /> }

        <div className="header">
          <Logo />
          <Topologies topologies={this.state.topologies} currentTopology={this.state.currentTopology} />
        </div>

        <Nodes nodes={this.state.nodes} highlightedNodeIds={this.state.highlightedNodeIds}
          highlightedEdgeIds={this.state.highlightedEdgeIds} detailsWidth={detailsWidth}
          selectedNodeId={this.state.selectedNodeId} topMargin={topMargin}
          topologyId={this.state.currentTopologyId} />

        <Sidebar>
          <TopologyOptions options={this.state.currentTopologyOptions}
            topologyId={this.state.currentTopologyId}
            activeOptions={this.state.activeTopologyOptions} />
          <Status errorUrl={this.state.errorUrl} topology={this.state.currentTopology}
            topologiesLoaded={this.state.topologiesLoaded}
            websocketClosed={this.state.websocketClosed} />
        </Sidebar>

        <div className="footer">
          {versionString}&nbsp;&nbsp;
          <a href="https://gitreports.com/issue/weaveworks/scope" target="_blank">Report an issue</a>
        </div>
      </div>
    );
  }
}
