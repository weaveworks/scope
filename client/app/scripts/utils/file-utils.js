// adapted from https://github.com/NYTimes/svg-crowbar
import _ from 'lodash';

const doctype = '<?xml version="1.0" standalone="no"?><!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">';
const prefix = {
  xmlns: 'http://www.w3.org/2000/xmlns/',
  xlink: 'http://www.w3.org/1999/xlink',
  svg: 'http://www.w3.org/2000/svg'
};

function setInlineStyles(svg, emptySvgDeclarationComputed) {
  function explicitlySetStyle(element) {
    const cSSStyleDeclarationComputed = getComputedStyle(element);
    let value;
    let computedStyleStr = '';
    _.each(cSSStyleDeclarationComputed, key => {
      value = cSSStyleDeclarationComputed.getPropertyValue(key);
      if (value !== emptySvgDeclarationComputed.getPropertyValue(key)) {
        computedStyleStr += key + ':' + value + ';';
      }
    });
    element.setAttribute('style', computedStyleStr);
  }

  function traverse(obj) {
    const tree = [];

    function visit(node) {
      if (node && node.hasChildNodes()) {
        let child = node.firstChild;
        while (child) {
          if (child.nodeType === 1 && child.nodeName !== 'SCRIPT') {
            tree.push(child);
            visit(child);
          }
          child = child.nextSibling;
        }
      }
    }

    tree.push(obj);
    visit(obj);
    return tree;
  }

  // hardcode computed css styles inside svg
  const allElements = traverse(svg);
  let i = allElements.length;
  while (i--) {
    explicitlySetStyle(allElements[i]);
  }
  // set font
  svg.setAttribute('style', 'font-family: "Roboto", sans-serif;');
}

function download(source) {
  let filename = 'untitled';

  if (source.id) {
    filename = source.id;
  } else if (source.class) {
    filename = source.class;
  } else if (window.document.title) {
    filename = window.document.title.replace(/[^a-z0-9]/gi, '-').toLowerCase();
  }

  const url = window.URL.createObjectURL(new Blob(source.source,
    {'type': 'text\/xml'}
  ));

  const a = document.createElement('a');
  document.body.appendChild(a);
  a.setAttribute('class', 'svg-crowbar');
  a.setAttribute('download', filename + '.svg');
  a.setAttribute('href', url);
  a.style.display = 'none';
  a.click();

  setTimeout(function() {
    window.URL.revokeObjectURL(url);
  }, 10);
}

function getSVG(doc, emptySvgDeclarationComputed) {
  const svg = document.getElementById('nodes-chart-canvas');

  svg.setAttribute('version', '1.1');

  // removing attributes so they aren't doubled up
  svg.removeAttribute('xmlns');
  svg.removeAttribute('xlink');

  // These are needed for the svg
  if (!svg.hasAttributeNS(prefix.xmlns, 'xmlns')) {
    svg.setAttributeNS(prefix.xmlns, 'xmlns', prefix.svg);
  }

  if (!svg.hasAttributeNS(prefix.xmlns, 'xmlns:xlink')) {
    svg.setAttributeNS(prefix.xmlns, 'xmlns:xlink', prefix.xlink);
  }

  setInlineStyles(svg, emptySvgDeclarationComputed);

  const source = (new XMLSerializer()).serializeToString(svg);

  return {
    class: svg.getAttribute('class'),
    id: svg.getAttribute('id'),
    childElementCount: svg.childElementCount,
    source: [doctype + source]
  };
}

function cleanup() {
  const crowbarElements = document.querySelectorAll('.svg-crowbar');

  [].forEach.call(crowbarElements, function(el) {
    el.parentNode.removeChild(el);
  });
}

export function saveGraph() {
  window.URL = (window.URL || window.webkitURL);

  // add empty svg element
  const emptySvg = window.document.createElementNS(prefix.svg, 'svg');
  window.document.body.appendChild(emptySvg);
  const emptySvgDeclarationComputed = getComputedStyle(emptySvg);

  const svgSource = getSVG(document, emptySvgDeclarationComputed);
  download(svgSource);

  cleanup();
}
