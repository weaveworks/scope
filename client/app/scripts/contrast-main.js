require('font-awesome-webpack');
require('../styles/contrast.less');

import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/app.js';

ReactDOM.render(<App base="/contrast.html" />, document.getElementById('app'));
