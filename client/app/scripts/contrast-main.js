require('font-awesome-webpack');
require('../styles/contrast.less');
require('../images/favicon.ico');

import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/app.js';

ReactDOM.render(<App />, document.getElementById('app'));
