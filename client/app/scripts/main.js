require('font-awesome-webpack');
require('../styles/main.less');

const React = require('react');
const ReactDOM = require('react-dom');

const App = require('./components/app.js');

ReactDOM.render(
  <App/>,
  document.getElementById('app'));
