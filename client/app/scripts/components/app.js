import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';

import Logo from './logo';
import Footer from './footer.js';
import Sidebar from './sidebar.js';
import HelpPanel from './help-panel';
import Search from './search';
import Status from './status.js';
import Topologies from './topologies.js';
import TopologyOptions from './topology-options.js';
import { getApiDetails, getTopologies } from '../utils/web-api-utils';
import { focusSearch, pinNextMetric, hitBackspace, hitEnter, hitEsc, unpinMetric,
  selectMetric, toggleHelp, toggleGridMode } from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import GridModeSelector from './grid-mode-selector';
import MetricSelector from './metric-selector';
import NetworkSelector from './networks-selector';
import { getRouter } from '../utils/router-utils';
import DebugToolbar, { showingDebugToolbar,
  toggleDebugToolbar } from './debug-toolbar.js';
import { getUrlState } from '../utils/router-utils';
import { getActiveTopologyOptions } from '../utils/topology-utils';

const BACKSPACE_KEY_CODE = 8;
const ENTER_KEY_CODE = 13;
const ESC_KEY_CODE = 27;
const keyPressLog = debug('scope:app-key-press');

class App extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onKeyPress = this.onKeyPress.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
  }

  componentDidMount() {
    window.addEventListener('keypress', this.onKeyPress);
    window.addEventListener('keyup', this.onKeyUp);

    getRouter(this.props.dispatch, this.props.urlState).start({hashbang: true});
    if (!this.props.routeSet) {
      // dont request topologies when already done via router
      getTopologies(this.props.activeTopologyOptions, this.props.dispatch);
    }
    getApiDetails(this.props.dispatch);
  }

  componentWillUnmount() {
    window.removeEventListener('keypress', this.onKeyPress);
    window.removeEventListener('keyup', this.onKeyUp);
  }

  onKeyUp(ev) {
    const { showingTerminal } = this.props;
    keyPressLog('onKeyUp', 'keyCode', ev.keyCode, ev);

    // don't get esc in onKeyPress
    if (ev.keyCode === ESC_KEY_CODE) {
      this.props.dispatch(hitEsc());
    } else if (ev.keyCode === ENTER_KEY_CODE) {
      this.props.dispatch(hitEnter());
    } else if (ev.keyCode === BACKSPACE_KEY_CODE) {
      this.props.dispatch(hitBackspace());
    } else if (ev.code === 'KeyD' && ev.ctrlKey && !showingTerminal) {
      toggleDebugToolbar();
      this.forceUpdate();
    }
  }

  onKeyPress(ev) {
    const { dispatch, searchFocused } = this.props;
    //
    // keyup gives 'key'
    // keypress gives 'char'
    // Distinction is important for international keyboard layouts where there
    // is often a different {key: char} mapping.
    //
    if (!searchFocused) {
      keyPressLog('onKeyPress', 'keyCode', ev.keyCode, ev);
      const char = String.fromCharCode(ev.charCode);
      if (char === '<') {
        dispatch(pinNextMetric(-1));
      } else if (char === '>') {
        dispatch(pinNextMetric(1));
      } else if (char === 't' || char === 'g') {
        dispatch(toggleGridMode());
      } else if (char === 'q') {
        dispatch(unpinMetric());
        dispatch(selectMetric(null));
      } else if (char === '/') {
        ev.preventDefault();
        dispatch(focusSearch());
      } else if (char === '?') {
        dispatch(toggleHelp());
      }
    }
  }

  render() {
    const { gridMode, showingDetails, showingHelp, showingMetricsSelector,
      showingNetworkSelector } = this.props;
    const isIframe = window !== window.top;

    return (
      <div className="app">
        {showingDebugToolbar() && <DebugToolbar />}

        {showingHelp && <HelpPanel />}

        {showingDetails && <Details />}

        <div className="header">
          <div className="logo">
            {!isIframe && <svg width="100%" height="100%" viewBox="0 0 1089 217">
              <Logo />
            </svg>}
          </div>
          <Search />
          <Topologies />
          <GridModeSelector />
        </div>

        <Nodes />

        <Sidebar classNames={gridMode ? 'sidebar-gridmode' : ''}>
          {showingMetricsSelector && !gridMode && <MetricSelector />}
          {showingNetworkSelector && !gridMode && <NetworkSelector />}
          <Status />
          <TopologyOptions />
        </Sidebar>

        <Footer />
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    activeTopologyOptions: getActiveTopologyOptions(state),
    gridMode: state.get('gridMode'),
    routeSet: state.get('routeSet'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    showingDetails: state.get('nodeDetails').size > 0,
    showingHelp: state.get('showingHelp'),
    showingMetricsSelector: state.get('availableCanvasMetrics').count() > 0,
    showingNetworkSelector: state.get('availableNetworks').count() > 0,
    showingTerminal: state.get('controlPipes').size > 0,
    urlState: getUrlState(state)
  };
}


export default connect(
  mapStateToProps
)(App);
