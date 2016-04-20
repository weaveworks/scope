require('../styles/main.less');
require('../images/favicon.ico');

import React from 'react';
import ReactDOM from 'react-dom';

import { TerminalApp } from './components/terminal-app.js';

ReactDOM.render(<TerminalApp />, document.getElementById('app'));
