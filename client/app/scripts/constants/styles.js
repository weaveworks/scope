
export const DETAILS_PANEL_WIDTH = 420;

export const DETAILS_PANEL_MARGINS = {
  top: 24,
  bottom: 48,
  right: 36
};

export const DETAILS_PANEL_OFFSET = 8;

export const CANVAS_MARGINS = {
  top: 160,
  left: 40,
  right: 40,
  bottom: 0,
};

// Node shapes
export const NODE_SHAPE_HIGHLIGHT_RADIUS = 70;
export const NODE_SHAPE_BORDER_RADIUS = 50;
export const NODE_SHAPE_SHADOW_RADIUS = 45;
export const NODE_SHAPE_DOT_RADIUS = 10;
// NOTE: This value represents the node unit radius (in pixels). Since zooming is
// controlled at the top level now, this renormalization would be obsolete (i.e.
// value 1 could be used instead), if it wasn't for the following factors:
//   1. `dagre` library only works with integer coordinates,
//      so >> 1 value is used to increase layout precision.
//   2. Fonts don't behave nicely (especially on Firefox) if they
//      are given on a small unit scale as foreign objects in SVG.
export const NODE_BASE_SIZE = 100;

// Node details table constants
export const NODE_DETAILS_TABLE_CW = {
  XS: '32px',
  S: '50px',
  M: '70px',
  L: '85px',
  XL: '120px',
  XXL: '140px',
  XXXL: '170px',
};

export const NODE_DETAILS_TABLE_COLUMN_WIDTHS = {
  count: NODE_DETAILS_TABLE_CW.XS,
  container: NODE_DETAILS_TABLE_CW.XS,
  docker_container_created: NODE_DETAILS_TABLE_CW.XXXL,
  docker_container_restart_count: NODE_DETAILS_TABLE_CW.M,
  docker_container_state_human: NODE_DETAILS_TABLE_CW.XXXL,
  docker_container_uptime: NODE_DETAILS_TABLE_CW.L,
  docker_cpu_total_usage: NODE_DETAILS_TABLE_CW.M,
  docker_memory_usage: NODE_DETAILS_TABLE_CW.M,
  open_files_count: NODE_DETAILS_TABLE_CW.M,
  pid: NODE_DETAILS_TABLE_CW.S,
  port: NODE_DETAILS_TABLE_CW.S,
  ppid: NODE_DETAILS_TABLE_CW.S,
  process_cpu_usage_percent: NODE_DETAILS_TABLE_CW.M,
  process_memory_usage_bytes: NODE_DETAILS_TABLE_CW.M,
  threads: NODE_DETAILS_TABLE_CW.M,

  // e.g. details panel > pods
  kubernetes_ip: NODE_DETAILS_TABLE_CW.XL,
  kubernetes_state: NODE_DETAILS_TABLE_CW.M,

  // weave connections
  weave_connection_connection: NODE_DETAILS_TABLE_CW.XXL,
  weave_connection_state: NODE_DETAILS_TABLE_CW.L,
  weave_connection_info: NODE_DETAILS_TABLE_CW.XL,
};

export const NODE_DETAILS_TABLE_XS_LABEL = {
  count: '#',
  // TODO: consider changing the name of this field on the BE
  container: '#',
};

// TODO: Make this variable
export const resourcesLayers = [{
  topologyId: 'hosts',
  horizontalPadding: 0,
  verticalPadding: 10,
  frameHeight: 200,
  withCapacity: true,
}, {
  topologyId: 'containers',
  horizontalPadding: 0,
  verticalPadding: 7.5,
  frameHeight: 150,
  withCapacity: false,
}, {
  topologyId: 'processes',
  horizontalPadding: 0,
  verticalPadding: 5,
  frameHeight: 100,
  withCapacity: false,
}];

export const layersDefs = {
  hosts: {
    parentTopologyId: null,
    verticalPadding: 10,
    frameHeight: 200,
    withCapacity: true,
  },
  containers: {
    parentTopologyId: 'hosts',
    verticalPadding: 7.5,
    frameHeight: 150,
    withCapacity: false,
  },
  processes: {
    parentTopologyId: 'containers',
    verticalPadding: 5,
    frameHeight: 100,
    withCapacity: false,
  },
};
