const reqwest = require('reqwest');

const AppActions = require('../actions/app-actions');

const WS_URL = window.WS_URL || 'ws://' + location.host;


let socket;
let reconnectTimer = 0;
let currentUrl = null;
let updateFrequency = '5s';
let topologyTimer = 0;
let apiDetailsTimer = 0;

function createWebsocket(topologyUrl) {
  if (socket) {
    socket.onclose = null;
    socket.close();
  }

  socket = new WebSocket(WS_URL + topologyUrl + '/ws?t=' + updateFrequency);

  socket.onclose = function() {
    clearTimeout(reconnectTimer);
    socket = null;

    reconnectTimer = setTimeout(function() {
      createWebsocket(topologyUrl);
    }, 5000);
  };

  socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    if (msg.add || msg.remove || msg.update) {
      AppActions.receiveNodesDelta(msg);
    }
  };

  currentUrl = topologyUrl;
}


const TOPOLOGIES = [
  {
    'name': 'Applications',
    'url': '/api/topology/applications',
    'stats': {
      'node_count': 12,
      'nonpseudo_node_count': 10,
      'edge_count': 13
    },
    'sub_topologies': [
      {
        'name': 'by name',
        'url': '/api/topology/applications-grouped'
      }
    ]
  },
  {
    'name': 'Containers',
    'url': '/api/topology/containers',
    'grouped_url': '/api/topology/containers-grouped',
    'stats': {
      'node_count': 2,
      'nonpseudo_node_count': 1,
      'edge_count': 2
    }
  },
  {
    'name': 'Hosts',
    'url': '/api/topology/hosts',
    'stats': {
      'node_count': 2,
      'nonpseudo_node_count': 1,
      'edge_count': 2
    }
  }
];


function getTopologies() {
  clearTimeout(topologyTimer);
  reqwest('/api/topology', function() {
    // injecting static topos
    AppActions.receiveTopologies(TOPOLOGIES);
    topologyTimer = setTimeout(getTopologies, 10000);
  });
}

function getNodeDetails(topologyUrl, nodeId) {
  if (topologyUrl && nodeId) {
    const url = [topologyUrl, nodeId].join('/');
    reqwest(url, function(res) {
      AppActions.receiveNodeDetails(res.node);
    });
  }
}

function getApiDetails() {
  clearTimeout(apiDetailsTimer);
  reqwest('/api', function(res) {
    AppActions.receiveApiDetails(res);
    apiDetailsTimer = setTimeout(getApiDetails, 10000);
  });
}

module.exports = {
  getNodeDetails: getNodeDetails,

  getTopologies: getTopologies,

  getApiDetails: getApiDetails,

  getNodesDelta: function(topologyUrl) {
    if (topologyUrl && topologyUrl !== currentUrl) {
      createWebsocket(topologyUrl);
    }
  }
};

