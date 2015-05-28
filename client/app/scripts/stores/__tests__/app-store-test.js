

describe('AppStore', function() {
  const ActionTypes = require('../../constants/action-types');
  let AppStore;
  let registeredCallback;

  // actions

  const ClickTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1'
  };

  const ClickGroupingAction = {
    type: ActionTypes.CLICK_GROUPING,
    grouping: 'grouped'
  };

  const ReceiveTopologiesAction = {
    type: ActionTypes.RECEIVE_TOPOLOGIES,
    topologies: [{
      url: '/topo1',
      grouped_url: '/topo1grouped',
      name: 'Topo1'
    }]
  };

  beforeEach(function() {
    AppStore = require('../app-store');
    registeredCallback = AppStore.registeredCallback;
  });

  // topology tests

  it('init with no topologies', function() {
    const topos = AppStore.getTopologies();
    expect(topos.length).toBe(0);
    expect(AppStore.getCurrentTopology()).toBeUndefined();
  });

  it('get current topology', function() {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);

    expect(AppStore.getTopologies().length).toBe(1);
    expect(AppStore.getCurrentTopology().name).toBe('Topo1');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1');
  });

  it('get grouped topology', function() {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickGroupingAction);

    expect(AppStore.getTopologies().length).toBe(1);
    expect(AppStore.getCurrentTopology().name).toBe('Topo1');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1grouped');
  });

});