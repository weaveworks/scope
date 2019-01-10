/* eslint no-return-assign: "off", react/jsx-no-bind: "off" */
import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';
import { debounce } from 'lodash';
import { Terminal as Term } from 'xterm';
import * as fit from 'xterm/lib/addons/fit/fit';

import { closeTerminal } from '../actions/app-actions';
import { getNeutralColor } from '../utils/color-utils';
import { setDocumentTitle } from '../utils/title-utils';
import { getPipeStatus, deletePipe, doResizeTty, getWebsocketUrl, basePath } from '../utils/web-api-utils';

const log = debug('scope:terminal');

const DEFAULT_COLS = 80;
const DEFAULT_ROWS = 24;
// Unicode points can be used in html and document.title
// html shorthand codes (&times;) don't work in document.title.
const TIMES = '\u00D7';
const MDASH = '\u2014';

const reconnectTimerInterval = 2000;


function ab2str(buf) {
  // http://stackoverflow.com/questions/17191945/conversion-between-utf-8-arraybuffer-and-string
  const encodedString = String.fromCharCode.apply(null, new Uint8Array(buf));
  const decodedString = decodeURIComponent(escape(encodedString));
  return decodedString;
}

function openNewWindow(url, bcr, minWidth = 200) {
  const screenLeft = window.screenX || window.screenLeft;
  const screenTop = window.screenY || window.screenTop;
  const popoutWindowToolbarHeight = 51;
  // TODO replace this stuff w/ looking up bounding box.
  const windowOptions = {
    height: bcr.height - popoutWindowToolbarHeight,
    left: screenLeft + bcr.left,
    location: 'no',
    top: screenTop + (window.outerHeight - window.innerHeight) + bcr.top,
    width: Math.max(minWidth, bcr.width),
  };

  const windowOptionsString = Object.keys(windowOptions)
    .map(k => `${k}=${windowOptions[k]}`)
    .join(',');

  window.open(url, '', windowOptionsString);
}


class Terminal extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.reconnectTimeout = null;
    this.resizeTimeout = null;

    this.state = {
      cols: DEFAULT_COLS,
      connected: false,
      detached: false,
      rows: DEFAULT_ROWS,
    };

    this.handleCloseClick = this.handleCloseClick.bind(this);
    this.handlePopoutTerminal = this.handlePopoutTerminal.bind(this);
    this.saveInnerFlexRef = this.saveInnerFlexRef.bind(this);
    this.saveNodeRef = this.saveNodeRef.bind(this);
    this.handleResize = this.handleResize.bind(this);
    this.handleResizeDebounced = debounce(this.handleResize, 500);
  }

  createWebsocket(term) {
    const socket = new WebSocket(`${getWebsocketUrl()}/api/pipe/${this.getPipeId()}`);
    socket.binaryType = 'arraybuffer';

    getPipeStatus(this.getPipeId(), this.props.dispatch);

    socket.onopen = () => {
      clearTimeout(this.reconnectTimeout);
      log('socket open to', getWebsocketUrl());
      this.setState({connected: true});
    };

    socket.onclose = () => {
      //
      // componentWillUnmount has called close and tidied up! don't try and do it again here
      // (setState etc), its too late.
      //
      if (!this.socket) {
        return;
      }
      this.socket = null;
      const wereConnected = this.state.connected;
      if (this.isComponentMounted) {
        // Calling setState on an unmounted component will throw a warning.
        // `connected` will get set to false by `componentWillUnmount`.
        this.setState({connected: false});
      }
      if (this.term && this.props.pipe.get('status') !== 'PIPE_DELETED') {
        if (wereConnected) {
          this.createWebsocket(term);
        } else {
          this.reconnectTimeout = setTimeout(
            this.createWebsocket.bind(this, term),
            reconnectTimerInterval
          );
        }
      }
    };

    socket.onerror = (err) => {
      log('socket error', err);
    };

    socket.onmessage = (event) => {
      log('pipe data', event.data.size);
      const input = ab2str(event.data);
      term.write(input);
    };

    this.socket = socket;
  }

  componentWillReceiveProps(nextProps) {
    if (this.props.connect !== nextProps.connect && nextProps.connect) {
      this.mountTerminal();
    }
    // Close the terminal window immediately when the pipe is deleted.
    if (nextProps.pipe.get('status') === 'PIPE_DELETED') {
      this.props.dispatch(closeTerminal(this.getPipeId()));
    }
  }

  componentDidMount() {
    this.isComponentMounted = true;
    if (this.props.connect) {
      this.mountTerminal();
    }
  }

  mountTerminal() {
    Term.applyAddon(fit);
    this.term = new Term({
      convertEol: !this.props.pipe.get('raw'),
      cursorBlink: true,
      //
      // Some linux systems fail to render 'monospace' on `<canvas>` correctly:
      // https://github.com/xtermjs/xterm.js/issues/1170
      // `theme.fontFamilies.monospace` doesn't provide many options so we add
      // some here that are very common. The alternative _might_ be to bundle Roboto-Mono
      //
      fontFamily: '"Roboto Mono", "Courier New", "Courier", monospace',
      // `theme.fontSizes.tiny` (`"12px"`) is a string and we need an int here.
      fontSize: 12,
      scrollback: 10000,
    });

    this.term.open(this.innerFlex);
    this.term.focus();

    this.term.on('data', (data) => {
      if (this.socket) {
        this.socket.send(data);
      }
    });

    this.term.on('resize', ({ cols, rows }) => {
      const resizeTtyControl = this.props.pipe.get('resizeTtyControl');
      if (resizeTtyControl) {
        doResizeTty(this.getPipeId(), resizeTtyControl, cols, rows);
      }
      this.setState({ cols, rows });
    });

    this.createWebsocket(this.term);

    window.addEventListener('resize', this.handleResizeDebounced);

    this.resizeTimeout = setTimeout(() => {
      this.handleResize();
    }, 10);
  }

  componentWillUnmount() {
    this.isComponentMounted = false;
    this.setState({connected: false});
    log('cwu terminal');

    clearTimeout(this.reconnectTimeout);
    clearTimeout(this.resizeTimeout);

    window.removeEventListener('resize', this.handleResizeDebounced);

    if (this.term) {
      log('destroy terminal');
      this.term.blur();
      this.term.destroy();
      this.term = null;
    }

    if (!this.state.detached) {
      deletePipe(this.getPipeId());
    }

    if (this.socket) {
      log('close socket');
      this.socket.close();
      this.socket = null;
    }
  }

  componentDidUpdate() {
    if (!this.isEmbedded()) {
      setDocumentTitle(this.getTitle());
    }
  }

  handleCloseClick(ev) {
    ev.preventDefault();
    this.props.dispatch(closeTerminal(this.getPipeId()));
  }

  handlePopoutTerminal(ev) {
    ev.preventDefault();
    const paramString = JSON.stringify(this.props);
    this.props.dispatch(closeTerminal(this.getPipeId()));
    this.setState({detached: true});

    const bcr = this.node.getBoundingClientRect();
    openNewWindow(`${basePath(window.location.pathname)}/terminal.html#!/state/${paramString}`, bcr);
  }

  handleResize() {
    this.term.fit();
  }

  isEmbedded() {
    return (this.props.embedded !== false);
  }

  getPipeId() {
    return this.props.pipe.get('id');
  }

  getTitle() {
    const nodeName = this.props.title || 'n/a';
    return `Terminal ${nodeName} ${MDASH}
      ${this.state.cols}${TIMES}${this.state.rows}`;
  }

  getTerminalHeader() {
    const light = this.props.statusBarColor || getNeutralColor();
    const style = {
      backgroundColor: light,
    };
    return (
      <div className="terminal-header" style={style}>
        <div className="terminal-header-tools">
          <span
            title="Open in new browser window"
            className="terminal-header-tools-item"
            onClick={this.handlePopoutTerminal}>
          Pop out
          </span>
          <i
            title="Close"
            className="terminal-header-tools-item-icon fa fa-times"
            onClick={this.handleCloseClick} />
        </div>
        {this.getControlStatusIcon()}
        <span className="terminal-header-title">{this.getTitle()}</span>
      </div>
    );
  }

  getStatus() {
    if (!this.state.connected) {
      return (
        <h3>Connecting...</h3>
      );
    }

    return (
      <h3>Connected</h3>
    );
  }

  getTerminalStatusBar() {
    const style = {
      backgroundColor: this.props.statusBarColor || getNeutralColor(),
      opacity: this.state.connected ? 0 : 0.9
    };
    return (
      <div className="terminal-status-bar hideable" style={style}>
        {this.getStatus()}
      </div>
    );
  }

  saveNodeRef(ref) {
    this.node = ref;
  }

  saveInnerFlexRef(ref) {
    this.innerFlex = ref;
  }

  render() {
    const innerFlexStyle = {
      opacity: this.state.connected ? 1 : 0.8,
      overflow: 'hidden',
    };
    const innerClassName = classNames('terminal-inner hideable', {
      'terminal-inactive': !this.state.connected
    });

    return (
      <div className="terminal-wrapper" ref={this.saveNodeRef}>
        {this.isEmbedded() && this.getTerminalHeader()}
        <div className={innerClassName} style={innerFlexStyle} ref={this.saveInnerFlexRef} />
        {this.getTerminalStatusBar()}
      </div>
    );
  }
  getControlStatusIcon() {
    const icon = this.props.controlStatus && this.props.controlStatus.get('control').icon;
    return (
      <i
        style={{marginRight: '8px', width: '14px'}}
        className={classNames('fa', {[icon]: icon})}
      />
    );
  }
}

function mapStateToProps(state, ownProps) {
  const controlStatus = state.get('controlPipes').find(pipe =>
    pipe.get('nodeId') === ownProps.pipe.get('nodeId'));
  return { controlStatus };
}

Terminal.defaultProps = {
  connect: true
};

export default connect(mapStateToProps)(Terminal);
