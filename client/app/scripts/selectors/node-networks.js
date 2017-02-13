import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { fromJS, Map as makeMap, List as makeList } from 'immutable';


const extractNodeNetworksValue = (node) => {
  if (node.has('metadata')) {
    const networks = node.get('metadata')
      .find(field => field.get('id') === 'docker_container_networks');
    return networks && networks.get('value');
  }
  return null;
};

// TODO: Move this setting of networks as toplevel node field to backend,
// to not rely on field IDs here. should be determined by topology implementer.
export const nodeNetworksSelector = createMapSelector(
  [
    state => state.get('nodes').map(extractNodeNetworksValue),
  ],
  (networksValue) => {
    if (!networksValue) {
      return makeList();
    }
    return fromJS(networksValue.split(', ').map(network => ({
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

// NOTE: Don't use this selector directly in mapStateToProps
// as it would get called too many times.
export const selectedNetworkNodesIdsSelector = createSelector(
  [
    state => state.get('networkNodes'),
    state => state.get('selectedNetwork'),
  ],
  (networkNodes, selectedNetwork) => networkNodes.get(selectedNetwork, makeMap())
);
