const React = require('react');
const mui = require('material-ui');

const Logo = require('./logo');
const AppStore = require('../stores/app-store');
const Sidebar = require('./sidebar.js');
const Status = require('./status.js');
const Topologies = require('./topologies.js');
const TopologyOptions = require('./topology-options.js');
const WebapiUtils = require('../utils/web-api-utils');
const AppActions = require('../actions/app-actions');
const Details = require('./details');
const Nodes = require('./nodes');
const RouterUtils = require('../utils/router-utils');

const ThemeManager = new mui.Styles.ThemeManager();

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


const App = React.createClass({

  getInitialState: function() {
    return getStateFromStores();
  },

  componentDidMount: function() {
    AppStore.on(AppStore.CHANGE_EVENT, this.onChange);
    window.addEventListener('keyup', this.onKeyPress);

    RouterUtils.getRouter().start({hashbang: true});
    if (!AppStore.isRouteSet()) {
      // dont request topologies when already done via router
      WebapiUtils.getTopologies(AppStore.getActiveTopologyOptions());
    }
    WebapiUtils.getApiDetails();
  },

  onChange: function() {
    this.setState(getStateFromStores());
  },

  onKeyPress: function(ev) {
    if (ev.keyCode === ESC_KEY_CODE) {
      AppActions.hitEsc();
    }
  },

  getChildContext: function() {
    return {
      muiTheme: ThemeManager.getCurrentTheme()
    };
  },

  render: function() {
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
  },

  childContextTypes: {
    muiTheme: React.PropTypes.object
  }
});

module.exports = App;
