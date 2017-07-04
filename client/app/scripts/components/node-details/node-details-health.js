import React from 'react';
import { Map as makeMap, List as makeList } from 'immutable';

import ShowMore from '../show-more';
import NodeDetailsHealthOverflow from './node-details-health-overflow';
import NodeDetailsHealthLinkItem from './node-details-health-link-item';
import CloudFeature from '../cloud-feature';

export default class NodeDetailsHealth extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      expanded: false
    };
    this.handleClickMore = this.handleClickMore.bind(this);
  }

  handleClickMore() {
    const expanded = !this.state.expanded;
    this.setState({expanded});
  }

  render() {
    const {
      metrics = makeList(),
      metricLinks = makeMap(),
      unattachedLinks = makeMap(),
      topologyId,
    } = this.props;

    const primeCutoff = metrics.length > 3 && !this.state.expanded ? 2 : metrics.length;
    const primeMetrics = metrics.slice(0, primeCutoff);
    const overflowMetrics = metrics.slice(primeCutoff);
    const showOverflow = overflowMetrics.length > 0 && !this.state.expanded;
    const flexWrap = showOverflow || !this.state.expanded ? 'nowrap' : 'wrap';
    const justifyContent = showOverflow || !this.state.expanded ? 'space-around' : 'flex-start';
    const notShown = overflowMetrics.length;

    return (
      <div className="node-details-health" style={{flexWrap, justifyContent}}>
        <div className="node-details-health-wrapper">
          {primeMetrics.map(item => <CloudFeature alwaysShow key={item.id}>
            <NodeDetailsHealthLinkItem
              {...item}
              links={metricLinks}
              topologyId={topologyId}
            />
          </CloudFeature>)}
          {showOverflow && <NodeDetailsHealthOverflow
            items={overflowMetrics}
            handleClick={this.handleClickMore}
          />}
        </div>
        <ShowMore
          handleClick={this.handleClickMore} collection={this.props.metrics}
          expanded={this.state.expanded} notShown={notShown} hideNumber />

        <div className="node-details-health-wrapper">
          {Object.keys(unattachedLinks).map(id => <CloudFeature alwaysShow key={id}>
            <NodeDetailsHealthLinkItem
              withoutGraph
              {...unattachedLinks[id]}
              links={unattachedLinks}
              topologyId={topologyId}
              />
          </CloudFeature>)}
        </div>
      </div>
    );
  }
}
