var d3 = require('d3');
var _ = require('lodash');
var EventEmitter = require('events').EventEmitter;


var MAX_NODES = 100;

function determineAllEdges(nodes) {
    var edges = [],
        edgeIds = {};

    _.each(nodes, function(node) {
        _.each(node.adjacency, function(adjacent) {
            var edge = [node.id, adjacent],
                edgeId = edge.join('-');

            if (!edgeIds[edgeId]) {
                edges.push({
                    id: edgeId,
                    value: 1,
                    source: nodes[edge[0]].index,
                    target: nodes[edge[1]].index
                });
                edgeIds[edgeId] = true;
            }
        });
    });

    return edges;
}

function chordChart() {

    var preview = false,
        margin = {top: 0, right: 0, bottom: 0, left: 0},
        width = 0,
        height = 0,
        outerRadius = 0,
        innerRadius = 0;

    var dispatcher = new EventEmitter();

    var formatPercent = d3.format(".1%");

    var arc = d3.svg.arc();

    var layout = d3.layout.chord();

    var path = d3.svg.chord();

    var color = d3.scale.category20();

    function chart(selection) {

        selection.each(function(data) {

            if (_.size(data.nodes) > MAX_NODES) {
                console.log('too many nodes for matrix');
                return;
            }

            width = Math.min(width, height);
            height = Math.min(width, height);

            if (preview) {
                outerRadius = Math.min(width, height) / 2 - 4,
                innerRadius = outerRadius - 4;
            } else {
                outerRadius = Math.min(width, height) / 2 - 10,
                innerRadius = outerRadius - 24;
            }

            arc
                .innerRadius(innerRadius)
                .outerRadius(outerRadius);

            path
                .radius(innerRadius);

            // Select the svg element, if it exists.
            var svg = d3.select(this)
                .selectAll("svg")
                .data([data]);

            svg.enter()
                .append("svg")
                .append("g")
                    .attr("class", "circle")
                .append("circle");

            svg
                .attr('width', width)
                .attr('height', height)

            svg.select('.circle')
                .attr("transform", "translate(" + width / 2 + "," + height / 2 + ")");

            svg.select('circle')
                .attr("r", outerRadius);

            var matrix = [],
                nodes = _.values(data.nodes),
                n = nodes.length,
                tooDense = n > width;

            // Compute index per node.
            nodes.forEach(function(node, i) {
                node.index = i;
                node.count = 0;
                matrix[i] = d3.range(n).map(function(j) {return 0;});
            });

            // Convert links to matrix; count character occurrences.
            var links = determineAllEdges(data.nodes);

            links.forEach(function(link) {
                matrix[link.source][link.target] = link.value / links.length;
            });

            layout
                .padding(preview ? .03 : .02)
                .sortGroups(d3.descending)
                .sortChords(d3.ascending);

            // Compute the chord layout.
            layout.matrix(matrix);

            // Add a group per node.
            var group = svg.select('.circle').selectAll(".group")
                .data(layout.groups);

            var groupEnter = group.enter().append("g")
                    .attr("class", "group");

            if (!preview) {
                // Add a mouseover title and interactivity
                groupEnter
                    .on('click', click)
                    .on("mouseout", mouseout)
                    .on("mouseover", mouseover)
                    .append("title");

                group.select('title').text(function(d, i) {
                    return nodes[i].label_major + ": " + formatPercent(d.value) + " of source edges";
                });
            }

            // Add the group arc.
            groupEnter.append("path")
                .attr('class', 'arc');

            var groupPath = group.select('path.arc');

            groupPath
                .attr("id", function(d, i) {
                    return preview ? '' : "group" + d.index;
                })
                .attr("d", arc)
                .style("fill", function(d, i) { return color(nodes[d.index].label_major); });

            if (!preview) {
                // Add a text label.
                groupEnter.append("text")
                    .attr("x", 6)
                    .attr("dy", 15)
                    .append("textPath")
                        .attr('class', 'textpath');

                group.select('.textpath')
                    .attr("xlink:href", function(d, i) {
                        return "#group" + i;
                    })
                    .text(function(d, i) { return nodes[i].label_major; });

                // Remove the labels that don't fit.
                group.select('text')
                    .classed('hide', false);

                group.select('text')
                    .classed('hide', function(d, i) {
                        return groupPath[0][i].getTotalLength() / 2 - 16 < this.getComputedTextLength();
                    });
            }

            group.exit()
                .remove();

            // Add the chords.
            var chord = svg.select('.circle').selectAll(".chord")
                .data(layout.chords);

            var chordEnter = chord.enter()
                .append("path")
                    .attr("class", "chord")
                    .on("mouseout", mouseoutChord)
                    .on("mouseover", mouseoverChord)
                .append("title");

            chord.exit()
                .remove();

            chord
                .style("fill", function(d) { return color(nodes[d.source.index].label_major); })
                .attr("d", path);

            // Add an elaborate mouseover title for each chord.
            chord.select('title')
                .text(function(d) {
                    var source = nodes[d.source.index],
                        target = nodes[d.target.index];

                  return source.label_major
                  + (source.label_minor ? "/" : "")
                  + (source.label_minor ? source.label_minor : "")
                  + " → " + target.label_major
                  + (target.label_minor ? "/" : "")
                  + (target.label_minor ? target.label_minor : "")
                  + ": " + formatPercent(d.source.value)
                  + "\n" + target.label_major
                  + (target.label_minor ? "/" : "")
                  + (target.label_minor ? target.label_minor : "")
                  + " → " + source.label_major
                  + (source.label_minor ? "/" : "")
                  + (source.label_minor ? source.label_minor : "")
                  + ": " + formatPercent(d.target.value);
                });

            var mouseOutTimer = 0;

            function mouseover(d, i) {
                clearTimeout(mouseOutTimer);
                chord.classed("fade", function(p) {
                    return p.source.index != i
                        && p.target.index != i;
                });
            }

            function mouseout(d, i) {
                clearTimeout(mouseOutTimer);
                mouseOutTimer = setTimeout(function() {
                    chord.classed("fade", function(p) {
                        return false;
                    });
                }, 500);
            }

            function mouseoverChord(d) {
                clearTimeout(mouseOutTimer);
                chord.classed("fade", function(p) {
                    return p.source.index != d.source.index;
                });
            }

            function mouseoutChord(d, i) {
                clearTimeout(mouseOutTimer);
                mouseOutTimer = setTimeout(function() {
                    chord.classed("fade", function(p) {
                        return false;
                    });
                }, 500);
            }


            function click(p) {
                dispatcher.emit('node.click', nodes[p.index].id);
            }
        });
    }

    chart.create = function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    };

    chart.update = _.throttle(function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    }, 500);

    chart.on = function(event, callback) {
        dispatcher.on(event, callback);
        return chart;
    };

    chart.preview = function(_) {
        if (!arguments.length) return preview;
        preview = _;
        return chart;
    };

    chart.width = function(_) {
        if (!arguments.length) return width;
        width = _;
        return chart;
    };

    chart.height = function(_) {
        if (!arguments.length) return height;
        height = _;
        return chart;
    };

    return chart;
}

module.exports = chordChart;