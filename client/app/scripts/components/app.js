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
  selectMetric, toggleHelp, setOptionKeyDown } from '../actions/app-actions';
import Details from './details';
import Nodes from './nodes';
import MetricSelector from './metric-selector';
import RawPipeDialog from './raw-pipe-dialog';
import EmbeddedTerminal from './embedded-terminal';
import { getRouter } from '../utils/router-utils';
import DebugToolbar, { showingDebugToolbar,
  toggleDebugToolbar } from './debug-toolbar.js';
import { getUrlState } from '../utils/router-utils';
import { getActiveTopologyOptions } from '../utils/topology-utils';

const BACKSPACE_KEY_CODE = 8;
const ENTER_KEY_CODE = 13;
const ESC_KEY_CODE = 27;
const ALT_KEY_CODE = 18;
const keyPressLog = debug('scope:app-key-press');

class App extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onKeyPress = this.onKeyPress.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
    this.onKeyDown = this.onKeyDown.bind(this);
  }

  componentDidMount() {
    window.addEventListener('keypress', this.onKeyPress);
    window.addEventListener('keyup', this.onKeyUp);
    window.addEventListener('keydown', this.onKeyDown);

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
    window.removeEventListener('keydown', this.onKeyDown);
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
    } else if (ev.keyCode === ALT_KEY_CODE) {
      this.props.dispatch(setOptionKeyDown(false));
    }
  }

  onKeyDown(ev) {
    keyPressLog('onKeyDown', 'keyCode', ev.keyCode, ev);
    if (ev.keyCode === ALT_KEY_CODE) {
      this.props.dispatch(setOptionKeyDown(true));
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
    const { showingDetails, showingHelp, showingMetricsSelector,
      showingTerminal, showingRawPipe } = this.props;

    return (
      <div className="app">
        {showingDebugToolbar() && <DebugToolbar />}

        {showingHelp && <HelpPanel />}

        {showingDetails && <Details />}

        {showingTerminal && <EmbeddedTerminal />}

        {showingRawPipe && <RawPipeDialog />}

        <div className="header">
          <div className="logo">
            <svg width="100%" height="100%" viewBox="0 0 1089 217">
              <Logo />
            </svg>
          </div>
          <Search />
          <Topologies />
        </div>

        <Nodes />

        <Sidebar>
          <Status />
          {showingMetricsSelector && <MetricSelector />}
          <TopologyOptions />
        </Sidebar>

        <Footer />
      </div>
    );
  }
}

function mapStateToProps(state) {
  const showingRawPipe = state.get('controlPipes').toIndexedSeq().getIn([0, 'rawPipeTemplate']);
  return {
    activeTopologyOptions: getActiveTopologyOptions(state),
    routeSet: state.get('routeSet'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    showingDetails: state.get('nodeDetails').size > 0,
    showingHelp: state.get('showingHelp'),
    showingMetricsSelector: state.get('availableCanvasMetrics').count() > 0,
    showingTerminal: !showingRawPipe && state.get('controlPipes').size > 0,
    urlState: getUrlState(state),
    showingRawPipe
  };
}

export default connect(
  mapStateToProps
)(App);
