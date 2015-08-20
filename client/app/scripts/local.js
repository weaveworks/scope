// This file is an entrypoint for development,
// see main.js for the real entrypoint

// Inject websocket url to dev backend
window.WS_URL = 'ws://' + location.hostname + ':4040';

require('./main');
