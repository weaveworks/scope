/* eslint no-return-assign: "off", react/jsx-no-bind: "off" */
import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';
import { debounce } from 'lodash';
import Term from 'xterm';

import { clickCloseTerminal } from '../actions/app-actions';
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

function terminalCellSize(wrapperNode) {
  // Badly guess the width/height of the row.
  let characterWidth = 20;
  let characterHeight = 20;

  // Now try and measure the first row we find.
  const subjectRow = wrapperNode.querySelector('.terminal .xterm-rows div');
  if (!subjectRow) {
    log("ERROR: Couldn't find first row, resizing might not work very well.");
  } else {
    const rowDisplay = subjectRow.style.display;
    const contentBuffer = subjectRow.innerHTML;

    subjectRow.innerHTML = 'W';
    subjectRow.style.display = 'inline';
    characterWidth = subjectRow.getBoundingClientRect().width;
    subjectRow.style.display = rowDisplay;
    characterHeight = parseInt(subjectRow.offsetHeight, 10);
    subjectRow.innerHTML = contentBuffer;
  }

  log('Calculated (charWidth, charHeight) sizes in px: ', characterWidth, characterHeight);
  return {characterWidth, characterHeight};
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
      connected: false,
      detached: false,
      rows: DEFAULT_ROWS,
      cols: DEFAULT_COLS,
      characterWidth: 0,
      characterHeight: 0
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
  }

  componentDidMount() {
    this.isComponentMounted = true;
    if (this.props.connect) {
      this.mountTerminal();
    }
  }

  mountTerminal() {
    this.term = new Term({
      cols: this.state.cols,
      rows: this.state.rows,
      convertEol: !this.props.pipe.get('raw'),
      cursorBlink: true,
      scrollback: 10000,
    });

    this.term.open(this.innerFlex);
    this.term.on('data', (data) => {
      if (this.socket) {
        this.socket.send(data);
      }
    });

    this.createWebsocket(this.term);

    const {characterWidth, characterHeight} = terminalCellSize(this.term.element);

    window.addEventListener('resize', this.handleResizeDebounced);

    this.resizeTimeout = setTimeout(() => {
      this.setState({
        characterWidth,
        characterHeight
      });
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
    this.props.dispatch(clickCloseTerminal(this.getPipeId()));
  }

  handlePopoutTerminal(ev) {
    ev.preventDefault();
    const paramString = JSON.stringify(this.props);
    this.props.dispatch(clickCloseTerminal(this.getPipeId()));
    this.setState({detached: true});

    const bcr = this.node.getBoundingClientRect();
    const minWidth = (this.state.characterWidth * 80) + (8 * 2);
    openNewWindow(`${basePath(window.location.pathname)}/terminal.html#!/state/${paramString}`, bcr, minWidth);
  }

  handleResize() {
    // scrollbar === 16px
    const width = this.innerFlex.clientWidth - (2 * 8) - 16;
    const height = this.innerFlex.clientHeight - (2 * 8);
    const cols = Math.floor(width / this.state.characterWidth);
    const rows = Math.floor(height / this.state.characterHeight);

    const resizeTtyControl = this.props.pipe.get('resizeTtyControl');
    if (resizeTtyControl) {
      doResizeTty(this.getPipeId(), resizeTtyControl, cols, rows)
        .then(() => this.setState({cols, rows}));
    } else if (!this.props.pipe.get('raw')) {
      this.setState({cols, rows});
    }
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
          <span
            title="Close"
            className="terminal-header-tools-item-icon fa fa-close"
            onClick={this.handleCloseClick} />
        </div>
        {this.getControlStatusIcon()}
        <span className="terminal-header-title">{this.getTitle()}</span>
      </div>
    );
  }

  getStatus() {
    if (this.props.pipe.get('status') === 'PIPE_DELETED') {
      return (
        <div>
          <h3>Connection Closed</h3>
          <div className="termina-status-bar-message">
            The connection to this container has been closed.
            <div className="link" onClick={this.handleCloseClick}>Close terminal</div>
          </div>
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
    const innerStyle = {
      width: (this.state.cols + 2) * this.state.characterWidth
    };
    const innerClassName = classNames('terminal-inner hideable', {
      'terminal-inactive': !this.state.connected
    });

    return (
      <div className="terminal-wrapper" ref={this.saveNodeRef}>
        {this.isEmbedded() && this.getTerminalHeader()}
        <div className={innerClassName} style={innerFlexStyle} ref={this.saveInnerFlexRef}>
          <div style={innerStyle} />
        </div>
        {this.getTerminalStatusBar()}
      </div>
    );
  }
  getControlStatusIcon() {
    const icon = this.props.controlStatus && this.props.controlStatus.get('control').icon;
    return (
      <span
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
