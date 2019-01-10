
// Cap the number of layers in the resource view to this constant. The reason why we have
// this constant is not just about the style, but also helps us build the selectors.
export const RESOURCE_VIEW_MAX_LAYERS = 3;

// TODO: Consider fetching these from the backend.
export const TOPOLOGIES_WITH_CAPACITY = ['hosts'];

// TODO: These too should ideally be provided by the backend. Currently, we are showing
// the same layers for all the topologies, because their number is small, but later on
// we might be interested in fully customizing the layers' hierarchy per topology.
export const RESOURCE_VIEW_LAYERS = {
  containers: ['hosts', 'containers', 'processes'],
  hosts: ['hosts', 'containers', 'processes'],
  processes: ['hosts', 'containers', 'processes'],
};

// TODO: These are all the common metrics that appear across all the current resource view
// topologies. The reason for taking them only is that we want to get meaningful data for all
// the layers. These should be taken directly from the backend report, but as their info is
// currently only contained in the nodes data, it would be hard to determine them before all
// the nodes for all the layers have been loaded, so we'd need to change the routing logic
// since the requirement is that one these is always pinned in the resource view.
export const RESOURCE_VIEW_METRICS = [
  { id: 'host_cpu_usage_percent', label: 'CPU' },
  { id: 'host_mem_usage_bytes', label: 'Memory' },
];
