import debug from 'debug';
import React from 'react';
import ReactDOM from 'react-dom';
import classNames from 'classnames';

import { clickCloseTerminal } from '../actions/app-actions';
import { getNeutralColor } from '../utils/color-utils';
import { setDocumentTitle } from '../utils/title-utils';
import { getPipeStatus, basePath } from '../utils/web-api-utils';
import Term from '../vendor/term.js';

const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';
const wsUrl = wsProto + '://' + location.host + basePath(location.pathname);
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

function terminalCellSize(wrapperNode, rows, cols) {
  const height = wrapperNode.clientHeight;

  // Guess the width of the row.
  let width = wrapperNode.clientWidth;
  // Now try and measure the first row we find.
  const firstRow = wrapperNode.querySelector('.terminal div');
  if (!firstRow) {
    log("ERROR: Couldn't find first row, resizing might not work very well.");
  } else {
    const rowDisplay = firstRow.style.display;
    firstRow.style.display = 'inline';
    width = firstRow.offsetWidth;
    firstRow.style.display = rowDisplay;
  }

  const pixelPerCol = width / cols;
  const pixelPerRow = height / rows;

  log('Caculated (col, row) sizes in px: ', pixelPerCol, pixelPerRow);
  return {pixelPerCol, pixelPerRow};
}

function openNewWindow(url, bcr, minWidth = 200) {
  const screenLeft = window.screenX || window.screenLeft;
  const screenTop = window.screenY || window.screenTop;
  const popoutWindowToolbarHeight = 51;
  // TODO replace this stuff w/ looking up bounding box.
  const windowOptions = {
    width: Math.max(minWidth, bcr.width),
    height: bcr.height - popoutWindowToolbarHeight,
    left: screenLeft + bcr.left,
    top: screenTop + (window.outerHeight - window.innerHeight) + bcr.top,
    location: 'no',
  };

  const windowOptionsString = Object.keys(windowOptions)
    .map((k) => k + '=' + windowOptions[k])
    .join(',');

  window.open(url, '', windowOptionsString);
}

export default class Terminal extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.reconnectTimeout = null;
    this.resizeTimeout = null;

    this.state = {
      connected: false,
      rows: DEFAULT_ROWS,
      cols: DEFAULT_COLS,
      pixelPerCol: 0,
      pixelPerRow: 0
    };
    this.handleCloseClick = this.handleCloseClick.bind(this);
    this.handlePopoutTerminal = this.handlePopoutTerminal.bind(this);
    this.handleResize = this.handleResize.bind(this);
  }

  createWebsocket(term) {
    const socket = new WebSocket(wsUrl + '/api/pipe/' + this.getPipeId());
    socket.binaryType = 'arraybuffer';

    getPipeStatus(this.getPipeId());

    socket.onopen = () => {
      clearTimeout(this.reconnectTimeout);
      log('socket open to', wsUrl);
      this.setState({connected: true});
    };

    socket.onclose = () => {
      log('socket closed');
      this.socket = null;
      const wereConnected = this.state.connected;
      this.setState({connected: false});
      if (this.term && this.props.pipe.status !== 'PIPE_DELETED') {
        if (wereConnected) {
          this.createWebsocket(term);
        } else {
          this.reconnectTimeout = setTimeout(
            this.createWebsocket.bind(this, term), reconnectTimerInterval);
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

  componentDidMount() {
    const component = this;

    this.term = new Term({
      cols: this.state.cols,
      rows: this.state.rows,
      convertEol: !this.props.raw
    });

    const innerNode = ReactDOM.findDOMNode(component.inner);
    this.term.open(innerNode);
    this.term.on('data', (data) => {
      if (this.socket) {
        this.socket.send(data);
      }
    });

    this.createWebsocket(this.term);

    const {pixelPerCol, pixelPerRow} = terminalCellSize(
      innerNode, this.state.rows, this.state.cols);

    window.addEventListener('resize', this.handleResize);

    this.resizeTimeout = setTimeout(() => {
      this.setState({
        pixelPerCol: pixelPerCol,
        pixelPerRow: pixelPerRow
      });
      this.handleResize();
    }, 10);
  }

  componentWillUnmount() {
    log('cwu terminal');

    clearTimeout(this.reconnectTimeout);
    clearTimeout(this.resizeTimeout);

    window.removeEventListener('resize', this.handleResize);

    if (this.term) {
      log('destroy terminal');
      this.term.destroy();
      this.term = null;
    }
    if (this.socket) {
      log('close socket');
      this.socket.close();
      this.socket = null;
    }
  }

  componentWillReceiveProps(nextProps) {
    const widthChanged = nextProps.containerWidth !== this.props.containerWidth;
    if (widthChanged) {
      this.handleResize();
    }
  }

  componentDidUpdate(prevProps, prevState) {
    const sizeChanged = (
      prevState.cols !== this.state.cols ||
      prevState.rows !== this.state.rows
    );
    if (sizeChanged) {
      this.term.resize(this.state.cols, this.state.rows);
    }
    if (!this.isEmbedded()) {
      setDocumentTitle(this.getTitle());
    }
  }

  handleCloseClick(ev) {
    ev.preventDefault();
    if (this.isEmbedded()) {
      clickCloseTerminal(this.getPipeId(), true);
    } else {
      window.close();
    }
  }

  handlePopoutTerminal(ev) {
    ev.preventDefault();
    const paramString = JSON.stringify(this.props);
    clickCloseTerminal(this.getPipeId());

    const bcr = ReactDOM.findDOMNode(this).getBoundingClientRect();
    const minWidth = this.state.pixelPerCol * 80 + (8 * 2);
    openNewWindow(`terminal.html#!/state/${paramString}`, bcr, minWidth);
  }

  handleResize() {
    const innerNode = ReactDOM.findDOMNode(this.innerFlex);
    const width = innerNode.clientWidth - (2 * 8);
    const height = innerNode.clientHeight - (2 * 8);
    const cols = Math.floor(width / this.state.pixelPerCol);
    const rows = Math.floor(height / this.state.pixelPerRow);
    this.setState({cols, rows});
  }

  isEmbedded() {
    return (this.props.embedded !== false);
  }

  getPipeId() {
    return this.props.pipe.id;
  }

  getTitle() {
    const nodeName = this.props.title || 'n/a';
    return `Terminal ${nodeName} ${MDASH}
      ${this.state.cols}${TIMES}${this.state.rows}`;
  }

  getTerminalHeader() {
    const style = {
      backgroundColor: this.props.titleBarColor || getNeutralColor()
    };
    return (
      <div className="terminal-header" style={style}>
          <div className="terminal-header-tools">
            <span className="terminal-header-tools-icon fa fa-external-link"
              onClick={this.handlePopoutTerminal} />
            <span className="terminal-header-tools-icon fa fa-close"
              onClick={this.handleCloseClick} />
          </div>
          <span className="terminal-header-title">{this.getTitle()}</span>
      </div>
    );
  }

  getStatus() {
    if (this.props.pipe.status === 'PIPE_DELETED') {
      return (
        <div>
          <h3>Connection Closed</h3>
          <p>
            The connection to this container has been closed.
            <div className="link" onClick={this.handleCloseClick}>Close terminal</div>
          </p>
        </div>
      );
    }

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

  render() {
    const innerStyle = {
      opacity: this.state.connected ? 1 : 0.8
    };
    const innerClassName = classNames('terminal-inner hideable', {
      'terminal-inactive': !this.state.connected
    });

    return (
      <div className="terminal-wrapper">
        {this.isEmbedded() && this.getTerminalHeader()}
        <div ref={(ref) => this.innerFlex = ref}
          className={innerClassName} style={innerStyle} >
          <div ref={(ref) => this.inner = ref} />
        </div>
        {this.getTerminalStatusBar()}
      </div>
    );
  }
}
