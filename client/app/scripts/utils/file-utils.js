// adapted from https://github.com/NYTimes/svg-crowbar
import { each } from 'lodash';

const doctype = '<?xml version="1.0" standalone="no"?><!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">';
const prefix = {
  svg: 'http://www.w3.org/2000/svg',
  xlink: 'http://www.w3.org/1999/xlink',
  xmlns: 'http://www.w3.org/2000/xmlns/'
};
const cssSkipValues = {
  '0px 0px': true,
  auto: true,
  pointer: true,
  visible: true
};

function setInlineStyles(svg, target, emptySvgDeclarationComputed) {
  function explicitlySetStyle(element, targetEl) {
    const cSSStyleDeclarationComputed = getComputedStyle(element);
    let computedStyleStr = '';
    each(cSSStyleDeclarationComputed, (key) => {
      const value = cSSStyleDeclarationComputed.getPropertyValue(key);
      if (value !== emptySvgDeclarationComputed.getPropertyValue(key) && !cssSkipValues[value]) {
        computedStyleStr += `${key}:${value};`;
      }
    });
    targetEl.setAttribute('style', computedStyleStr);
    targetEl.removeAttribute('data-reactid');
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

  // make sure logo shows up
  svg.setAttribute('class', 'exported');

  // hardcode computed css styles inside svg
  const allElements = traverse(svg);
  const allTargetElements = traverse(target);
  for (let i = allElements.length - 1; i >= 0; i -= 1) {
    explicitlySetStyle(allElements[i], allTargetElements[i]);
  }

  // set font
  target.setAttribute('style', 'font-family: Arial;');

  // set view box
  target.setAttribute('width', svg.clientWidth);
  target.setAttribute('height', svg.clientHeight);
}

function download(source, name) {
  let filename = 'untitled';

  if (name) {
    filename = name;
  } else if (window.document.title) {
    filename = `${window.document.title.replace(/[^a-z0-9]/gi, '-').toLowerCase()}-${(+new Date())}`;
  }

  const url = window.URL.createObjectURL(new Blob(
    source,
    { type: 'text/xml' }
  ));

  const a = document.createElement('a');
  document.body.appendChild(a);
  a.setAttribute('class', 'svg-crowbar');
  a.setAttribute('download', `${filename}.svg`);
  a.setAttribute('href', url);
  a.style.display = 'none';
  a.click();

  setTimeout(() => {
    window.URL.revokeObjectURL(url);
  }, 10);
}

function getSVGElement() {
  return document.getElementById('canvas');
}

function getSVG(doc, emptySvgDeclarationComputed) {
  const svg = getSVGElement();
  const target = svg.cloneNode(true);

  target.setAttribute('version', '1.1');

  // removing attributes so they aren't doubled up
  target.removeAttribute('xmlns');
  target.removeAttribute('xlink');

  // These are needed for the svg
  if (!target.hasAttributeNS(prefix.xmlns, 'xmlns')) {
    target.setAttributeNS(prefix.xmlns, 'xmlns', prefix.svg);
  }

  if (!target.hasAttributeNS(prefix.xmlns, 'xmlns:xlink')) {
    target.setAttributeNS(prefix.xmlns, 'xmlns:xlink', prefix.xlink);
  }

  setInlineStyles(svg, target, emptySvgDeclarationComputed);

  const source = (new XMLSerializer()).serializeToString(target);

  return [doctype + source];
}

function cleanup() {
  const crowbarElements = document.querySelectorAll('.svg-crowbar');

  [].forEach.call(crowbarElements, (el) => {
    el.parentNode.removeChild(el);
  });

  // hide embedded logo
  const svg = getSVGElement();
  svg.setAttribute('class', '');
}

export function saveGraph(filename) {
  window.URL = (window.URL || window.webkitURL);

  // add empty svg element
  const emptySvg = window.document.createElementNS(prefix.svg, 'svg');
  window.document.body.appendChild(emptySvg);
  const emptySvgDeclarationComputed = getComputedStyle(emptySvg);

  const svgSource = getSVG(document, emptySvgDeclarationComputed);
  download(svgSource, filename);

  cleanup();
}
