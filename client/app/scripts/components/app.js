import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import Logo from './logo';
import Footer from './footer';
import Sidebar from './sidebar';
import HelpPanel from './help-panel';
import TroubleshootingMenu from './troubleshooting-menu';
import Search from './search';
import Status from './status';
import Topologies from './topologies';
import TopologyOptions from './topology-options';
import CloudFeature from './cloud-feature';
import Overlay from './overlay';
import { getApiDetails } from '../utils/web-api-utils';
import {
  focusSearch,
  pinNextMetric,
  pinPreviousMetric,
  hitBackspace,
  hitEsc,
  unpinMetric,
  toggleHelp,
  setGraphView,
  setTableView,
  setResourceView,
  shutdown,
  setViewportDimensions,
  getTopologiesWithInitialPoll,
} from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import TimeTravel from './time-travel';
import TimeControl from './time-control';
import ViewModeSelector from './view-mode-selector';
import NetworkSelector from './networks-selector';
import DebugToolbar, { showingDebugToolbar, toggleDebugToolbar } from './debug-toolbar';
import { getRouter, getUrlState } from '../utils/router-utils';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import { availableNetworksSelector } from '../selectors/node-networks';
import {
  isResourceViewModeSelector,
  isTableViewModeSelector,
  isGraphViewModeSelector,
} from '../selectors/topology';
import { VIEWPORT_RESIZE_DEBOUNCE_INTERVAL } from '../constants/timer';
import {
  BACKSPACE_KEY_CODE,
  ESC_KEY_CODE,
} from '../constants/key-codes';

const keyPressLog = debug('scope:app-key-press');


class App extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.setViewportDimensions = this.setViewportDimensions.bind(this);
    this.handleResize = debounce(this.setViewportDimensions, VIEWPORT_RESIZE_DEBOUNCE_INTERVAL);

    this.saveAppRef = this.saveAppRef.bind(this);
    this.onKeyPress = this.onKeyPress.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
  }

  componentDidMount() {
    this.setViewportDimensions();
    window.addEventListener('resize', this.handleResize);
    window.addEventListener('keypress', this.onKeyPress);
    window.addEventListener('keyup', this.onKeyUp);

    getRouter(this.props.dispatch, this.props.urlState).start({hashbang: true});
    if (!this.props.routeSet || process.env.WEAVE_CLOUD) {
      // dont request topologies when already done via router.
      // If running as a component, always request topologies when the app mounts.
      this.props.dispatch(getTopologiesWithInitialPoll());
    }
    getApiDetails(this.props.dispatch);
  }

  componentWillUnmount() {
    window.addEventListener('resize', this.handleResize);
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
        dispatch(pinPreviousMetric());
        this.trackEvent('scope.metric.selector.pin.previous.keypress', {
          metricType: this.props.pinnedMetricType
        });
      } else if (char === '>') {
        dispatch(pinNextMetric());
        this.trackEvent('scope.metric.selector.pin.next.keypress', {
          metricType: this.props.pinnedMetricType
        });
      } else if (char === 'g') {
        dispatch(setGraphView());
        this.trackEvent('scope.layout.selector.keypress');
      } else if (char === 't') {
        dispatch(setTableView());
        this.trackEvent('scope.layout.selector.keypress');
      } else if (char === 'r') {
        dispatch(setResourceView());
        this.trackEvent('scope.layout.selector.keypress');
      } else if (char === 'q') {
        this.trackEvent('scope.metric.selector.unpin.keypress', {
          metricType: this.props.pinnedMetricType
        });
        dispatch(unpinMetric());
      } else if (char === '/') {
        ev.preventDefault();
        dispatch(focusSearch());
      } else if (char === '?') {
        dispatch(toggleHelp());
      }
    }
  }

  trackEvent(eventName, additionalProps = {}) {
    trackMixpanelEvent(eventName, {
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      ...additionalProps,
    });
  }

  setViewportDimensions() {
    if (this.appRef) {
      const { width, height } = this.appRef.getBoundingClientRect();
      this.props.dispatch(setViewportDimensions(width, height));
    }
  }

  saveAppRef(ref) {
    this.appRef = ref;
  }

  render() {
    const { isTableViewMode, isGraphViewMode, isResourceViewMode, showingDetails, showingHelp,
      showingNetworkSelector, showingTroubleshootingMenu, timeTravelTransitioning } = this.props;
    const isIframe = window !== window.top;

    return (
      <div className="scope-app" ref={this.saveAppRef}>
        {showingDebugToolbar() && <DebugToolbar />}

        {showingHelp && <HelpPanel />}

        {showingTroubleshootingMenu && <TroubleshootingMenu />}

        {showingDetails && <Details />}

        <div className="header">
          <CloudFeature>
            <TimeTravel />
          </CloudFeature>

          <div className="selectors">
            <div className="logo">
              {!isIframe && <svg width="100%" height="100%" viewBox="0 0 1089 217">
                <Logo />
              </svg>}
            </div>
            <Search />
            <Topologies />
            <ViewModeSelector />
            <TimeControl />
          </div>
        </div>

        <Nodes />

        <Sidebar classNames={isTableViewMode ? 'sidebar-gridmode' : ''}>
          {showingNetworkSelector && isGraphViewMode && <NetworkSelector />}
          {!isResourceViewMode && <Status />}
          {!isResourceViewMode && <TopologyOptions />}
        </Sidebar>

        <Footer />

        <Overlay faded={timeTravelTransitioning} />
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    currentTopology: state.get('currentTopology'),
    isResourceViewMode: isResourceViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isGraphViewMode: isGraphViewModeSelector(state),
    pinnedMetricType: state.get('pinnedMetricType'),
    routeSet: state.get('routeSet'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    showingDetails: state.get('nodeDetails').size > 0,
    showingHelp: state.get('showingHelp'),
    showingTroubleshootingMenu: state.get('showingTroubleshootingMenu'),
    showingNetworkSelector: availableNetworksSelector(state).count() > 0,
    showingTerminal: state.get('controlPipes').size > 0,
    topologyViewMode: state.get('topologyViewMode'),
    timeTravelTransitioning: state.get('timeTravelTransitioning'),
    urlState: getUrlState(state)
  };
}

export default connect(
  mapStateToProps
)(App);
