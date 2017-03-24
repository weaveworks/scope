import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { fromJS, List as makeList, Map as makeMap } from 'immutable';


const NETWORKS_ID = 'docker_container_networks';

// TODO: Move this setting of networks as toplevel node field to backend,
// to not rely on field IDs here. should be determined by topology implementer.
export const nodeNetworksSelector = createMapSelector(
  [
    state => state.get('nodes'),
  ],
  (node) => {
    const metadata = node.get('metadata', makeList());
    const networks = metadata.find(f => f.get('id') === NETWORKS_ID) || makeMap();
    const networkValues = networks.has('value') ? networks.get('value').split(', ') : [];

    return fromJS(networkValues.map(network => ({
      id: network, label: network, colorKey: network
    })));
  }
);

export const availableNetworksSelector = createSelector(
  [
    nodeNetworksSelector
  ],
  networksMap => networksMap.toList().flatten(true).toSet().toList()
    .sortBy(m => m.get('label'))
);

export const selectedNetworkNodesIdsSelector = createSelector(
  [
    nodeNetworksSelector,
    state => state.get('selectedNetwork'),
  ],
  (nodeNetworks, selectedNetworkId) => {
    const nodeIds = [];
    nodeNetworks.forEach((networks, nodeId) => {
      const networksIds = networks.map(n => n.get('id'));
      if (networksIds.contains(selectedNetworkId)) {
        nodeIds.push(nodeId);
      }
    });
    return fromJS(nodeIds);
  }
);
