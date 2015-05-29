require('font-awesome-webpack');
require('../styles/main.less');

const React = require('react');

const App = require('./components/app.js');

React.render(
  <App/>,
  document.getElementById('app'));
