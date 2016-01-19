import React from 'react';

import NodeDetailsHealthOverflow from './node-details-health-overflow';
import NodeDetailsHealthItem from './node-details-health-item';

export default class NodeDetailsHealth extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      expanded: false
    };
    this.handleClickMore = this.handleClickMore.bind(this);
  }

  handleClickMore(ev) {
    ev.preventDefault();
    const expanded = !this.state.expanded;
    this.setState({expanded});
  }

  render() {
    const metrics = this.props.metrics || [];
    const primeCutoff = metrics.length > 3 && !this.state.expanded ? 2 : metrics.length;
    const primeMetrics = metrics.slice(0, primeCutoff);
    const overflowMetrics = metrics.slice(primeCutoff);
    const showOverflow = overflowMetrics.length > 0 && !this.state.expanded;
    const showLess = this.state.expanded;
    const flexWrap = showOverflow || !this.state.expanded ? 'nowrap' : 'wrap';
    const justifyContent = showOverflow || !this.state.expanded ? 'space-around' : 'flex-start';

    return (
      <div className="node-details-health" style={{flexWrap, justifyContent}}>
          {primeMetrics.map(item => {
            return <NodeDetailsHealthItem key={item.id} item={item} />;
          })}
        {showOverflow && <NodeDetailsHealthOverflow items={overflowMetrics} handleClickMore={this.handleClickMore} />}
        {showLess && <div className="node-details-health-expand" onClick={this.handleClickMore}>show less</div>}
      </div>
    );
  }
}
