import React from 'react';

import Logo from './logo';
import AppStore from '../stores/app-store';
import Footer from './footer.js';
import Sidebar from './sidebar.js';
import Status from './status.js';
import Topologies from './topologies.js';
import TopologyOptions from './topology-options.js';
import { getApiDetails, getTopologies } from '../utils/web-api-utils';
import { hitEsc } from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import MetricSelector from './metric-selector';
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
    forceRelayout: AppStore.isForceRelayout(),
    highlightedEdgeIds: AppStore.getHighlightedEdgeIds(),
    highlightedNodeIds: AppStore.getHighlightedNodeIds(),
    hostname: AppStore.getHostname(),
    nodeDetails: AppStore.getNodeDetails(),
    nodes: AppStore.getNodes(),
    selectedNodeId: AppStore.getSelectedNodeId(),
    selectedMetric: AppStore.getSelectedMetric(),
    topologies: AppStore.getTopologies(),
    topologiesLoaded: AppStore.isTopologiesLoaded(),
    updatePaused: AppStore.isUpdatePaused(),
    updatePausedAt: AppStore.getUpdatePausedAt(),
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
    const {nodeDetails, controlPipe } = this.state;
    const showingDetails = nodeDetails.size > 0;
    const showingTerminal = controlPipe;
    // width of details panel blocking a view
    const detailsWidth = showingDetails ? 450 : 0;
    const topMargin = 100;

    return (
      <div className="app">
        {showingDebugToolbar() && <DebugToolbar />}
        {showingDetails && <Details nodes={this.state.nodes}
          controlStatus={this.state.controlStatus}
          details={this.state.nodeDetails} />}

        {showingTerminal && <EmbeddedTerminal
          pipe={this.state.controlPipe}
          nodeId={this.state.controlPipe.nodeId}
          details={this.state.nodeDetails} />}

        <div className="header">
          <div className="logo">
            <svg width="100%" height="100%" viewBox="0 0 1089 217">
              <Logo />
            </svg>
          </div>
          <Topologies topologies={this.state.topologies}
            currentTopology={this.state.currentTopology} />
        </div>

        <Nodes
          nodes={this.state.nodes}
          highlightedNodeIds={this.state.highlightedNodeIds}
          highlightedEdgeIds={this.state.highlightedEdgeIds}
          detailsWidth={detailsWidth}
          selectedNodeId={this.state.selectedNodeId}
          topMargin={topMargin}
          selectedMetric={this.state.selectedMetric}
          forceRelayout={this.state.forceRelayout}
          topologyOptions={this.state.activeTopologyOptions}
          topologyId={this.state.currentTopologyId} />

        <Sidebar>
          <MetricSelector selectedMetric={this.state.selectedMetric}/>
          <Status errorUrl={this.state.errorUrl} topology={this.state.currentTopology}
            topologiesLoaded={this.state.topologiesLoaded}
            websocketClosed={this.state.websocketClosed} />
          <TopologyOptions options={this.state.currentTopologyOptions}
            topologyId={this.state.currentTopologyId}
            activeOptions={this.state.activeTopologyOptions} />
        </Sidebar>

        <Footer {...this.state} />
      </div>
    );
  }
}
