require('font-awesome-webpack');
require('../styles/main.less');

import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/app.js';

ReactDOM.render(<App />, document.getElementById('app'));
