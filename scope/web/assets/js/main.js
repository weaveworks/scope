$(function() {
	var $graph = $('#graph'),
		$win = $(window),
		$hostDetail = $('#host-detail'),

        line = d3.svg.line()
        	.x(function(d) { return d.x; })
    		.y(function(d) { return d.y; }),

        slices = ['host-lb1', 'host-lb2', 'host-elastic1', 'host-elastic2',
		'host-elastic3', 'host-elastic4'],
        angle = Math.PI * 2 / slices.length,

		locations = {},
		scrollY = 0,
		phase0, phase1, phase2, phase3, centerX, centerY, radius, width;

	// bind data to svg elements

	d3.selectAll('#graph g.host')
		.each(function(d) {
			d3.select(this).datum(this.id);
		})
		.on('click', function(d) {
			showHostDetail(d);
		});

	d3.selectAll('#graph g.link-host')
		.each(function(d) {
			var parts = this.id.split('-'),
				hostB = 'host-' + parts.pop(),
				hostA = 'host-' + parts.pop();
			// assign hosts to links
			d3.select(this).datum([hostA, hostB]);
		});

	var showHostDetail = function(host) {
		d3.selectAll('#graph g.host')
			.classed('active', function(d) {
				return d === host;
			});

		$hostDetail.attr('src', 'assets/img/' + host + '.png');
	}

	var phaseHandler = function(phaseA, phaseB, ratio) {
		d3.selectAll('#graph g.host')
			.classed('faded', function(d) {
				return phaseA.get(d).hide && phaseB.get(d).hide;
			})
			.attr('transform', function(d) {
				var optsA = phaseA.get(d),
					optsB = phaseB.get(d),
					dx = (optsB.x - optsA.x) * ratio + optsA.x,
					dy = (optsB.y - optsA.y) * ratio + optsA.y;

				// caching locations for links
				locations[d] = {x: dx, y: dy};

	            return 'translate(' + dx + ',' + dy + ')';
	        });

		d3.selectAll('#graph g.link-host')
			.classed('faded', function(d) {
				var hostA = d[0],
					hostB = d[1];

				return phaseB.get(hostA).hide
					|| phaseB.get(hostB).hide;
			})
			.each(function(l) {
	            d3.select(this).selectAll('path').attr("d", function(d) {
	                return line([locations[l[0]], locations[l[1]]]);
	            });
			})
	}

	var initializePhases = function() {

		width = $graph.width(),
		centerX = width / 12 * 3;
		centerY = 700;
		radius = width / 4 - 80;

		phase0 = d3.map({
			'host-theinternet': {x: width / 12 * 1, y: 120, hide: false},
			'host-lb1': {x: width / 12 * 5, y: 120, hide: false},
			'host-app1': {x: width / 12 * 7, y: 180, hide: false},
			'host-app2': {x: width / 12 * 8, y: 60, hide: false},
			'host-lb2': {x: width / 2, y: 50, hide: true},
			'host-elastic4': {x: width / 2, y: 50, hide: true},
			'host-elastic1': {x: width / 12 * 10, y: 180, hide: false},
			'host-elastic2': {x: width / 12 * 11, y: 60, hide: false},
			'host-elastic3': {x: width / 2, y: 50, hide: true}
		});

		phase1 = d3.map({
			'host-theinternet': {x: width / 12 * 2, y: 120, hide: false},
			'host-lb1': {x: width / 12 * 4, y: 150, hide: false},
			'host-app1': {x: width / 12 * 6, y: 250, hide: false},
			'host-app2': {x: width / 12 * 8, y: 60, hide: false},
			'host-lb2': {x: width / 2, y: 50, hide: true},
			'host-elastic4': {x: width / 2, y: 50, hide: true},
			'host-elastic1': {x: width / 12 * 7, y: 200, hide: false},
			'host-elastic2': {x: width / 12 * 9, y: 150, hide: false},
			'host-elastic3': {x: width / 2, y: 50, hide: true}
		});

		phase2 = d3.map({
			'host-theinternet': {x: centerX, y: 150, hide: false},
			'host-lb1': {x: centerX, y: 250, hide: false},
			'host-app1': {x: width / 12 * 4, y: 300, hide: false},
			'host-app2': {x: width / 12 * 8, y: 60, hide: true},
			'host-lb2': {x: centerX, y: 500, hide: true},
			'host-elastic4': {x: centerX, y: 500, hide: true},
			'host-elastic1': {x: width / 12 * 6, y: 300, hide: false},
			'host-elastic2': {x: width / 12 * 7, y: 200, hide: false},
			'host-elastic3': {x: centerX, y: 500, hide: true}
		});

		phase3 = d3.map({
			'host-theinternet': {x: width / 12 * 3, y: 180, hide: false},
			// center
			'host-app1': {x: width / 12 * 3, y: centerY, hide: false},
			// radial
			'host-lb1': {x: centerX + radius * Math.sin(angle * 3), y: centerY + radius * Math.cos(angle * 3), hide: false},
			'host-lb2': {x: centerX + radius * Math.sin(angle * 4), y: centerY + radius * Math.cos(angle * 4), hide: false},
			'host-elastic4': {x: centerX + radius * Math.sin(angle * 5), y: centerY + radius * Math.cos(angle * 5), hide: false},
			'host-elastic1': {x: centerX + radius * Math.sin(angle * 0), y: centerY + radius * Math.cos(angle * 0), hide: false},
			'host-elastic2': {x: centerX + radius * Math.sin(angle * 1), y: centerY + radius * Math.cos(angle * 1), hide: false},
			'host-elastic3': {x: centerX + radius * Math.sin(angle * 2), y: centerY + radius * Math.cos(angle * 2), hide: false},
			'host-app2': {x: width / 2, y: centerY, hide: true}
		});
	};

    var animationHandler = function(e) {
		scrollY = $win.scrollTop();
		if (scrollY < 100) {
			scrollY = Math.max(scrollY, -10);
			phaseHandler(phase0, phase1, scrollY / 100);
		} else if (scrollY < 200) {
			phaseHandler(phase1, phase2, (scrollY - 100) / 100);
		} else if (scrollY < 300) {
			phaseHandler(phase2, phase3, (scrollY - 200) / 100);
		} else if (scrollY < 1000) {
			phaseHandler(phase2, phase3, 1);
		};
    };

    var resizeHandler = function() {
    	initializePhases();
    	animationHandler();
    };

    var onMouseOverVideo = function(e) {
    	var el = $(this).find('video').get(0);

    	el.playbackRate = 2;
	    el.play();
        window.ga && window.ga('send', 'event', 'video', 'start', el.src);
    };

    var onMouseOutVideo = function(e) {
    	var el = $(this).find('video').get(0);

    	el.pause();
    	$(el).data('count', 0); // reset play count
        window.ga && window.ga('send', 'event', 'video', 'stop', el.src);
    };

    $(document).scroll(animationHandler);
    $win.resize(resizeHandler);
	initializePhases();
    phaseHandler(phase0, phase0, 1);

    // video play on mouseover

    $('#featureswrap .row')
    	.on('mouseover', onMouseOverVideo)
    	.on('mouseout', onMouseOutVideo);
    $('video')
    	.on('ended', function(e) {
    		var $el = $(this),
    			count = $el.data('count') || 0;

    		this.currentTime = 0.1;
    		this.load();
	    	this.playbackRate = 2;
    		if (count < 5) {  // stop after 5 loops
	    		this.play();
	    		$el.data('count', count + 1);
    		}
    	});

});
