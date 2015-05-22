
describe('AppStore', function() {

	var ActionTypes = require('../../constants/action-types');
	var AppStore, registeredCallback;

	// actions

	var ClickTopologyAction = {
		type: ActionTypes.CLICK_TOPOLOGY,
		topologyId: 'topo1'
	};

	var ClickGroupingAction = {
		type: ActionTypes.CLICK_GROUPING,
		grouping: 'grouped'
	};

	var ReceiveTopologiesAction = {
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
		var topos = AppStore.getTopologies();
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