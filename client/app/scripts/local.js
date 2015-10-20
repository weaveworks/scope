// This file is an entrypoint for development,
// see main.js for the real entrypoint

// Inject websocket url to dev backend
window.WS_PROTO = (location.protocol === 'https:' ? 'wss' : 'ws');
window.WS_URL = window.WS_PROTO + '://' + location.hostname + ':4040';

require('./main');
