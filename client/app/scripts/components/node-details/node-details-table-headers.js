import React from 'react';
import { defaultSortDesc, getTableColumnsStyles } from '../../utils/node-details-utils';
import { NODE_DETAILS_TABLE_CW, NODE_DETAILS_TABLE_XS_LABEL } from '../../constants/styles';


export default class NodeDetailsTableHeaders extends React.Component {
  handleClick(ev, headerId, currentSortedBy, currentSortedDesc) {
    ev.preventDefault();
    const header = this.props.headers.find(h => h.id === headerId);
    const sortedBy = header.id;
    const sortedDesc = sortedBy === currentSortedBy
      ? !currentSortedDesc : defaultSortDesc(header);
    this.props.onClick(sortedBy, sortedDesc);
  }

  render() {
    const { headers, sortedBy, sortedDesc } = this.props;
    const colStyles = getTableColumnsStyles(headers);
    return (
      <tr>
        {headers.map((header, index) => {
          const headerClasses = ['node-details-table-header', 'truncate'];
          const onClick = (ev) => {
            this.handleClick(ev, header.id, sortedBy, sortedDesc);
          };
          // sort by first metric by default
          const isSorted = header.id === sortedBy;
          const isSortedDesc = isSorted && sortedDesc;
          const isSortedAsc = isSorted && !isSortedDesc;

          if (isSorted) {
            headerClasses.push('node-details-table-header-sorted');
          }

          const style = colStyles[index];
          const label =
            (style.width === NODE_DETAILS_TABLE_CW.XS && NODE_DETAILS_TABLE_XS_LABEL[header.id]) ?
            NODE_DETAILS_TABLE_XS_LABEL[header.id] : header.label;

          return (
            <td className={headerClasses.join(' ')} style={style} title={header.label} key={header.id}>
              <div className="node-details-table-header-sortable" onClick={onClick}>
                {isSortedAsc
                  && <i className="node-details-table-header-sorter fa fa-caret-up" />}
                {isSortedDesc
                  && <i className="node-details-table-header-sorter fa fa-caret-down" />}
                {label}
              </div>
            </td>
          );
        })}
      </tr>
    );
  }
}
