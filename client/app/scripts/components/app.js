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
import EmbeddedTerminal from './embedded-terminal';
import { getRouter } from '../utils/router-utils';
import { showingDebugToolbar, DebugToolbar } from './debug-toolbar.js';

const ESC_KEY_CODE = 27;

function getStateFromStores() {
  return {
    activeTopologyOptions: AppStore.getActiveTopologyOptions(),
    controlStatus: AppStore.getControlStatus(),
    controlPipe: AppStore.getControlPipe(),
    currentTopology: AppStore.getCurrentTopology(),
    currentTopologyId: AppStore.getCurrentTopologyId(),
    currentTopologyOptions: AppStore.getCurrentTopologyOptions(),
    errorUrl: AppStore.getErrorUrl(),
    highlightedEdgeIds: AppStore.getHighlightedEdgeIds(),
    highlightedNodeIds: AppStore.getHighlightedNodeIds(),
    hostname: AppStore.getHostname(),
    nodeDetails: AppStore.getNodeDetails(),
    nodes: AppStore.getNodes(),
    selectedNodeId: AppStore.getSelectedNodeId(),
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
    const showingDetails = this.state.nodeDetails.size > 0;
    const showingTerminal = this.state.controlPipe;
    const footer = `Version ${this.state.version} on ${this.state.hostname}`;
    // width of details panel blocking a view
    const detailsWidth = showingDetails ? 450 : 0;
    const topMargin = 100;

    return (
      <div className="app">
        {showingDebugToolbar() && <DebugToolbar />}
        {showingDetails && <Details nodes={this.state.nodes}
          controlStatus={this.state.controlStatus[this.state.selectedNodeId]}
          details={this.state.nodeDetails} />}

        {showingTerminal && <EmbeddedTerminal
          pipe={this.state.controlPipe}
          nodeId={this.state.controlPipe.nodeId}
          nodes={this.state.nodes} />}

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
          {footer}&nbsp;&nbsp;
          <a href="https://gitreports.com/issue/weaveworks/scope" target="_blank">Report an issue</a>
        </div>
      </div>
    );
  }
}
