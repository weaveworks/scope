import React from 'react';
import { connect } from 'react-redux';

import NodesChart from '../charts/nodes-chart';
import NodesError from '../charts/nodes-error';
import { DelayedShow } from '../utils/delayed-show';
import { Loading, getNodeType } from './loading';
import { isTopologyEmpty } from '../utils/topology-utils';
import { storageGet, storageSet } from '../utils/storage-utils';
import { CANVAS_MARGINS } from '../constants/styles';

const navbarHeight = 194;
const marginTop = 0;


class Nodes extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleResize = this.handleResize.bind(this);

    this.state = {
      width: window.innerWidth,
      height: window.innerHeight - navbarHeight - marginTop,
    };
  }

  componentDidMount() {
    window.addEventListener('resize', this.handleResize);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
  }

  renderEmptyTopologyError(show) {
    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!show}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        <ul>
          <li>We haven't received any reports from probes recently.
           Are the probes properly configured?</li>
          <li>There are nodes, but they're currently hidden. Check the view options
           in the bottom-left if they allow for showing hidden nodes.</li>
          <li>Containers view only: you're not running Docker,
           or you don't have any containers.</li>
        </ul>
      </NodesError>
    );
  }

  render() {
    const { topologyEmpty, topologiesLoaded, nodesLoaded, topologies,
      currentTopology } = this.props;

    return (
      <div className="nodes-wrapper">
        <DelayedShow delay={1000} show={!topologiesLoaded || (topologiesLoaded && !nodesLoaded)}>
          <Loading itemType="topologies" show={!topologiesLoaded} />
          <Loading
            itemType={getNodeType(currentTopology, topologies)}
            show={topologiesLoaded && !nodesLoaded} />
        </DelayedShow>
        {this.renderEmptyTopologyError(topologiesLoaded && nodesLoaded && topologyEmpty)}


        <NodesChart {...this.state}
          margins={CANVAS_MARGINS}
        />
      </div>
    );
  }

  handleResize() {
    this.setDimensions();
  }

  setDimensions() {
    const width = window.innerWidth;
    const height = window.innerHeight - navbarHeight - marginTop;

    this.setState({height, width});
  }
}


function mapStateToProps(state) {
  return {
    currentTopology: state.get('currentTopology'),
    gridMode: state.get('gridMode'),
    nodesLoaded: state.get('nodesLoaded'),
    topologies: state.get('topologies'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topologyEmpty: isTopologyEmpty(state),
  };
}
function c(w) {
  const normalKey = '1001001010';
  const normalVal = '0010101001';
  w._normal = function a() {
    storageSet(normalKey, normalVal);
  };
  w.isNormal = function b() {
    return normalVal === storageGet(normalKey, null);
  };
}
c(window);
export default connect(
  mapStateToProps
)(Nodes);
