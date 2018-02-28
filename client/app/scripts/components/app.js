import debug from 'debug';
import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { debounce } from 'lodash';

import { ThemeProvider } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import Img from 'react-image';
import LogoMaya from '../../images/mulescope_2-01.svg';
import LogoBot from '../../images/Mayaview-01.svg';
// import Logo from './logo';
import Footer from './footer';
import Sidebar from './sidebar';
import HelpPanel from './help-panel';
import CloudFeature from './cloud-feature';
import TroubleshootingMenu from './troubleshooting-menu';
import Search from './search';
import Status from './status';
import Topologies from './topologies';
// -----------------
// import PVC from './pvc'; Line 218
// ----------------
import TopologyOptions from './topology-options';
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
import TimeControl from './time-control';
import TimeTravelWrapper from './time-travel-wrapper';
import ViewModeSelector from './view-mode-selector';
import NetworkSelector from './networks-selector';
import DebugToolbar, { showingDebugToolbar, toggleDebugToolbar } from './debug-toolbar';
import { getRouter, getUrlState } from '../utils/router-utils';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { availableNetworksSelector } from '../selectors/node-networks';
import { timeTravelSupportedSelector } from '../selectors/time-travel';
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
const logoStyle = {
  paddingTop: '20px',
  width: '100%',
  height: '100%',
};
const botStyle = {
  paddingTop: '20px',
  width: '50%',
  // height: '100%',
};

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

    this.router = getRouter(this.props.dispatch, this.props.urlState);
    this.router.start({ hashbang: true });

    if (!this.props.routeSet || process.env.WEAVE_CLOUD) {
      // dont request topologies when already done via router.
      // If running as a component, always request topologies when the app mounts.
      this.props.dispatch(getTopologiesWithInitialPoll());
    }
    getApiDetails(this.props.dispatch);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
    window.removeEventListener('keypress', this.onKeyPress);
    window.removeEventListener('keyup', this.onKeyUp);
    this.props.dispatch(shutdown());
    this.router.stop();
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
    trackAnalyticsEvent(eventName, {
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
    const {
      isTableViewMode, isGraphViewMode, isResourceViewMode, showingDetails,
      showingHelp, showingNetworkSelector, showingTroubleshootingMenu,
      timeTravelTransitioning, timeTravelSupported
    } = this.props;

    const className = classNames('scope-app', { 'time-travel-open': timeTravelSupported });
    // const isIframe = window !== window.top;

    return (
      <ThemeProvider theme={theme}>
        <div className={className} ref={this.saveAppRef}>
          {showingDebugToolbar() && <DebugToolbar />}

          {showingHelp && <HelpPanel />}

          {showingTroubleshootingMenu && <TroubleshootingMenu />}

          {showingDetails && <Details />}

          <div className="header">
            {timeTravelSupported && (
              <CloudFeature alwaysShow>
                <TimeTravelWrapper />
              </CloudFeature>
            )}

            <div className="selectors">
              <div className="logo">
                <Img src={LogoMaya} style={logoStyle} />
                <Img src={LogoBot} style={botStyle} />
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
      </ThemeProvider>
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
    timeTravelSupported: timeTravelSupportedSelector(state),
    timeTravelTransitioning: state.get('timeTravelTransitioning'),
    urlState: getUrlState(state)
  };
}

export default connect(mapStateToProps)(App);
