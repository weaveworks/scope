import { NODE_DETAILS_TABLE_COLUMN_WIDTHS } from '../constants/styles';

export function isGenericTable(table) {
  return (table.type || (table.get && table.get('type'))) === 'multicolumn-table';
}

export function isPropertyList(table) {
  return (table.type || (table.get && table.get('type'))) === 'property-list';
}

export function isNumber(data) {
  return data && data.dataType && data.dataType === 'number';
}

/** Whether the value is considered numeric for sorting purposes. */
export function isNumeric(data) {
  return data && data.dataType && (data.dataType === 'number' || data.dataType === 'duration');
}

export function isIP(data) {
  return data && data.dataType && data.dataType === 'ip';
}

export function genericTableEntryKey(row, column) {
  const columnId = column.id || column.get('id');
  const rowId = row.id || row.get('id');
  return `${rowId}_${columnId}`;
}

export function defaultSortDesc(header) {
  return header && isNumber(header);
}

export function getTableColumnsStyles(headers) {
  return headers.map(header => ({
    textAlign: isNumber(header) ? 'right' : 'left',
    // More beauty hacking, ports and counts can only get
    // so big, free up WS for other longer fields like IPs!
    width: NODE_DETAILS_TABLE_COLUMN_WIDTHS[header.id]
  }));
}
