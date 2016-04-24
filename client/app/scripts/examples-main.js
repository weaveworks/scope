require('../styles/main.less');
require('../../node_modules/fixed-data-table/dist/fixed-data-table.css');
require('../images/favicon.ico');

import React from 'react';
import ReactDOM from 'react-dom';

import { Examples } from './components/examples.js';

ReactDOM.render(<Examples />, document.getElementById('app'));
