import debug from 'debug';
import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import Logo from './logo';
import AppStore from '../stores/app-store';
import Footer from './footer.js';
import Sidebar from './sidebar.js';
import HelpPanel from './help-panel';
import Status from './status.js';
import Topologies from './topologies.js';
import TopologyOptions from './topology-options.js';
import Plugins from './plugins.js';
import { getApiDetails, getTopologies } from '../utils/web-api-utils';
import { pinNextMetric, hitEsc, unpinMetric,
  selectMetric, toggleHelp } from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import MetricSelector from './metric-selector';
import EmbeddedTerminal from './embedded-terminal';
import { getRouter } from '../utils/router-utils';
import { showingDebugToolbar, toggleDebugToolbar,
  DebugToolbar } from './debug-toolbar.js';

const ESC_KEY_CODE = 27;
const keyPressLog = debug('scope:app-key-press');

/* make sure these can all be shallow-checked for equality for PureRenderMixin */
function getStateFromStores() {
  return {
    activeTopologyOptions: AppStore.getActiveTopologyOptions(),
    adjacentNodes: AppStore.getAdjacentNodes(AppStore.getSelectedNodeId()),
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
    pinnedMetric: AppStore.getPinnedMetric(),
    availableCanvasMetrics: AppStore.getAvailableCanvasMetrics(),
    nodeDetails: AppStore.getNodeDetails(),
    nodes: AppStore.getNodes(),
    showingHelp: AppStore.getShowingHelp(),
    selectedNodeId: AppStore.getSelectedNodeId(),
    selectedMetric: AppStore.getSelectedMetric(),
    topologies: AppStore.getTopologies(),
    topologiesLoaded: AppStore.isTopologiesLoaded(),
    topologyEmpty: AppStore.isTopologyEmpty(),
    updatePaused: AppStore.isUpdatePaused(),
    updatePausedAt: AppStore.getUpdatePausedAt(),
    version: AppStore.getVersion(),
    plugins: AppStore.getPlugins(),
    websocketClosed: AppStore.isWebsocketClosed()
  };
}

export default class App extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);
    this.onKeyPress = this.onKeyPress.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
    this.state = getStateFromStores();
  }

  componentDidMount() {
    AppStore.addListener(this.onChange);
    window.addEventListener('keypress', this.onKeyPress);
    window.addEventListener('keyup', this.onKeyUp);

    getRouter().start({hashbang: true});
    if (!AppStore.isRouteSet()) {
      // dont request topologies when already done via router
      getTopologies(AppStore.getActiveTopologyOptions());
    }
    getApiDetails();
  }

  componentWillUnmount() {
    window.removeEventListener('keypress', this.onKeyPress);
    window.removeEventListener('keyup', this.onKeyUp);
  }

  onChange() {
    this.setState(getStateFromStores());
  }

  onKeyUp(ev) {
    // don't get esc in onKeyPress
    if (ev.keyCode === ESC_KEY_CODE) {
      hitEsc();
    }
  }

  onKeyPress(ev) {
    //
    // keyup gives 'key'
    // keypress gives 'char'
    // Distinction is important for international keyboard layouts where there
    // is often a different {key: char} mapping.
    //
    keyPressLog('onKeyPress', 'keyCode', ev.keyCode, ev);
    const char = String.fromCharCode(ev.charCode);
    if (char === '<') {
      pinNextMetric(-1);
    } else if (char === '>') {
      pinNextMetric(1);
    } else if (char === 'q') {
      unpinMetric();
      selectMetric(null);
    } else if (char === 'd') {
      toggleDebugToolbar();
      this.forceUpdate();
    } else if (char === '?') {
      toggleHelp();
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

        {this.state.showingHelp && <HelpPanel />}

        {showingDetails && <Details nodes={this.state.nodes}
          controlStatus={this.state.controlStatus}
          details={this.state.nodeDetails} />}

        {showingTerminal && <EmbeddedTerminal
          pipe={this.state.controlPipe}
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
          topologyEmpty={this.state.topologyEmpty}
          adjacentNodes={this.state.adjacentNodes}
          topologyId={this.state.currentTopologyId} />

        <Sidebar>
          <Status errorUrl={this.state.errorUrl} topology={this.state.currentTopology}
            topologiesLoaded={this.state.topologiesLoaded}
            websocketClosed={this.state.websocketClosed} />
          {this.state.availableCanvasMetrics.count() > 0 && <MetricSelector
            availableCanvasMetrics={this.state.availableCanvasMetrics}
            pinnedMetric={this.state.pinnedMetric}
            selectedMetric={this.state.selectedMetric}
            />}
          <TopologyOptions options={this.state.currentTopologyOptions}
            topologyId={this.state.currentTopologyId}
            activeOptions={this.state.activeTopologyOptions} />
          <Plugins plugins={this.state.plugins} />
        </Sidebar>

        <Footer {...this.state} />
      </div>
    );
  }
}

reactMixin.onClass(App, PureRenderMixin);
