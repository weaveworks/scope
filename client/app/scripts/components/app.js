import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';

import Logo from './logo';
import Footer from './footer';
import Sidebar from './sidebar';
import HelpPanel from './help-panel';
import TroubleshootingMenu from './troubleshooting-menu';
import Search from './search';
import Status from './status';
import Topologies from './topologies';
import TopologyOptions from './topology-options';
import { getApiDetails, getTopologies } from '../utils/web-api-utils';
import {
  focusSearch,
  pinNextMetric,
  hitBackspace,
  hitEnter,
  hitEsc,
  unpinMetric,
  toggleHelp,
  setGraphView,
  setTableView,
  setResourceView,
  shutdown
} from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import ViewModeSelector from './view-mode-selector';
import NetworkSelector from './networks-selector';
import DebugToolbar, { showingDebugToolbar, toggleDebugToolbar } from './debug-toolbar';
import { getRouter, getUrlState } from '../utils/router-utils';
import { availableNetworksSelector } from '../selectors/node-networks';
import {
  activeTopologyOptionsSelector,
  isResourceViewModeSelector,
  isTableViewModeSelector,
  isGraphViewModeSelector,
} from '../selectors/topology';


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
    if (!this.props.routeSet || process.env.WEAVE_CLOUD) {
      // dont request topologies when already done via router.
      // If running as a component, always request topologies when the app mounts.
      getTopologies(this.props.activeTopologyOptions, this.props.dispatch, true);
    }
    getApiDetails(this.props.dispatch);
  }

  componentWillUnmount() {
    window.removeEventListener('keypress', this.onKeyPress);
    window.removeEventListener('keyup', this.onKeyUp);
    this.props.dispatch(shutdown());
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
    const { dispatch, searchFocused, showingTerminal } = this.props;
    //
    // keyup gives 'key'
    // keypress gives 'char'
    // Distinction is important for international keyboard layouts where there
    // is often a different {key: char} mapping.
    if (!searchFocused && !showingTerminal) {
      keyPressLog('onKeyPress', 'keyCode', ev.keyCode, ev);
      const char = String.fromCharCode(ev.charCode);
      if (char === '<') {
        dispatch(pinNextMetric(-1));
      } else if (char === '>') {
        dispatch(pinNextMetric(1));
      } else if (char === 'g') {
        dispatch(setGraphView());
      } else if (char === 't') {
        dispatch(setTableView());
      } else if (char === 'r') {
        dispatch(setResourceView());
      } else if (char === 'q') {
        dispatch(unpinMetric());
      } else if (char === '/') {
        ev.preventDefault();
        dispatch(focusSearch());
      } else if (char === '?') {
        dispatch(toggleHelp());
      }
    }
  }

  render() {
    const { isTableViewMode, isGraphViewMode, isResourceViewMode, showingDetails, showingHelp,
      showingNetworkSelector, showingTroubleshootingMenu } = this.props;
    const isIframe = window !== window.top;

    return (
      <div className="scope-app">
        {showingDebugToolbar() && <DebugToolbar />}

        {showingHelp && <HelpPanel />}

        {showingTroubleshootingMenu && <TroubleshootingMenu />}

        {showingDetails && <Details />}

        <div className="header">
          <div className="logo">
            {!isIframe && <svg width="100%" height="100%" viewBox="0 0 1089 217">
              <Logo />
            </svg>}
          </div>
          <Search />
          <Topologies />
          <ViewModeSelector />
        </div>

        <Nodes />

        <Sidebar classNames={isTableViewMode ? 'sidebar-gridmode' : ''}>
          {showingNetworkSelector && isGraphViewMode && <NetworkSelector />}
          {!isResourceViewMode && <Status />}
          {!isResourceViewMode && <TopologyOptions />}
        </Sidebar>

        <Footer />
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    activeTopologyOptions: activeTopologyOptionsSelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isGraphViewMode: isGraphViewModeSelector(state),
    routeSet: state.get('routeSet'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    showingDetails: state.get('nodeDetails').size > 0,
    showingHelp: state.get('showingHelp'),
    showingTroubleshootingMenu: state.get('showingTroubleshootingMenu'),
    showingNetworkSelector: availableNetworksSelector(state).count() > 0,
    showingTerminal: state.get('controlPipes').size > 0,
    urlState: getUrlState(state)
  };
}

export default connect(
  mapStateToProps
)(App);
