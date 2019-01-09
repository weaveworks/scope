import { GRAPH_VIEW_MODE, RESOURCE_VIEW_MODE } from './naming';


export const DETAILS_PANEL_WIDTH = 420;
export const DETAILS_PANEL_OFFSET = 8;
export const DETAILS_PANEL_MARGINS = {
  bottom: 48,
  right: 36,
  top: 24
};

// Resource view
export const RESOURCES_LAYER_TITLE_WIDTH = 200;
export const RESOURCES_LAYER_HEIGHT = 150;
export const RESOURCES_LAYER_PADDING = 10;
export const RESOURCES_LABEL_MIN_SIZE = 50;
export const RESOURCES_LABEL_PADDING = 10;

// Node shapes
export const UNIT_CLOUD_PATH = 'M-1.25 0.233Q-1.25 0.44-1.104 0.587-0.957 0.733-0.75 0.733H0.667Q'
  + '0.908 0.733 1.079 0.562 1.25 0.391 1.25 0.15 1.25-0.022 1.158-0.164 1.065-0.307 0.914-0.377q'
  + '0.003-0.036 0.003-0.056 0-0.276-0.196-0.472-0.195-0.195-0.471-0.195-0.206 0-0.373 0.115-0.167'
  + ' 0.115-0.244 0.299-0.091-0.081-0.216-0.081-0.138 0-0.236 0.098-0.098 0.098-0.098 0.236 0 0.098'
  + ' 0.054 0.179-0.168 0.039-0.278 0.175-0.109 0.136-0.109 0.312z';

// Node Cylinder shape
export const UNIT_CYLINDER_PATH = 'm -1 -1.25' // this line is responsible for adjusting place of the shape with respect to dot
  + 'a 1 0.4 0 0 0 2 0'
  + 'm -2 0'
  + 'v 1.8'
  + 'a 1 0.4 0 0 0 2 0'
  + 'v -1.8'
  + 'a 1 0.4 0 0 0 -2 0';

// Node Storage Sheet Shape
export const SHEET = 'm -1.2 -1.6 m 0.4 0 v 2.4 m -0.4 -2.4 v 2.4 h 2 v -2.4 z m 0 0.4 h 2';

// NOTE: This value represents the node unit radius (in pixels). Since zooming is
// controlled at the top level now, this renormalization would be obsolete (i.e.
// value 1 could be used instead), if it wasn't for the following factors:
//   1. `dagre` library only works with integer coordinates,
//      so >> 1 value is used to increase layout precision.
//   2. Fonts don't behave nicely (especially on Firefox) if they
//      are given on a small unit scale as foreign objects in SVG.
export const NODE_BASE_SIZE = 100;

// This value represents the upper bound on the number of control points along the graph edge
// curve. Any integer value >= 6 should result in valid edges, but generally the greater this
// value is, the nicer the edge bundling will be. On the other hand, big values would result
// in slower rendering of the graph.
export const EDGE_WAYPOINTS_CAP = 10;

export const CANVAS_MARGINS = {
  [GRAPH_VIEW_MODE]: {
    bottom: 150, left: 80, right: 80, top: 220
  },
  [RESOURCE_VIEW_MODE]: {
    bottom: 150, left: 210, right: 40, top: 200
  },
};

// Node details table constants
export const NODE_DETAILS_TABLE_CW = {
  L: '85px',
  M: '70px',
  // 6 chars wide with our current font choices, (pids can be 6, ports only 5).
  S: '56px',
  XL: '120px',
  XS: '32px',
  XXL: '140px',
  XXXL: '170px',
};

export const NODE_DETAILS_TABLE_COLUMN_WIDTHS = {
  container: NODE_DETAILS_TABLE_CW.XS,
  count: NODE_DETAILS_TABLE_CW.XS,
  docker_container_created: NODE_DETAILS_TABLE_CW.XXXL,
  docker_container_restart_count: NODE_DETAILS_TABLE_CW.M,
  docker_container_state_human: NODE_DETAILS_TABLE_CW.XXXL,
  docker_container_uptime: NODE_DETAILS_TABLE_CW.L,
  docker_cpu_total_usage: NODE_DETAILS_TABLE_CW.M,
  docker_memory_usage: NODE_DETAILS_TABLE_CW.M,
  // e.g. details panel > pods
  kubernetes_ip: NODE_DETAILS_TABLE_CW.XL,
  kubernetes_state: NODE_DETAILS_TABLE_CW.M,
  open_files_count: NODE_DETAILS_TABLE_CW.M,
  pid: NODE_DETAILS_TABLE_CW.S,
  port: NODE_DETAILS_TABLE_CW.S,
  // Label "Parent PID" needs more space
  ppid: NODE_DETAILS_TABLE_CW.M,
  process_cpu_usage_percent: NODE_DETAILS_TABLE_CW.M,

  process_memory_usage_bytes: NODE_DETAILS_TABLE_CW.M,
  threads: NODE_DETAILS_TABLE_CW.M,

  // weave connections
  weave_connection_connection: NODE_DETAILS_TABLE_CW.XXL,
  weave_connection_info: NODE_DETAILS_TABLE_CW.XL,
  weave_connection_state: NODE_DETAILS_TABLE_CW.L,
};

export const NODE_DETAILS_TABLE_XS_LABEL = {
  // TODO: consider changing the name of this field on the BE
  container: '#',
  count: '#',
};
