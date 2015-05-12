
PANDOC = pandoc

DOCS=install concepts rules eula
HTMLFILES=$(DOCS:%=docs/%.html)

all: $(HTMLFILES)

docs/%.html: docs/header.html docs/%.md
	$(PANDOC) -S -H docs/header.html -o $@ $(@:%.html=%.md)

.PHONY: clean
clean:
	rm -f $(HTMLFILES)

.PHONY: test
test:
	echo

.PHONY: screenshots
# needs to have ./scenario_screenshot running
hosts = app1 app2 dns1 elastic1 elastic2 elastic3 elastic4 lb1 lb2 theinternet
screenshots:
	for h in $(hosts); do webkit2png -H 1280 -W 1280 -F -o $$h --delay=5 --js='setTimeout(function(){$$("svg").addClass("hide")}, 3000);' http://localhost:4040/topology/all/all/all/hosts/$$h; done
	for h in $(hosts); do convert $$h-full.png -crop 410x850+850+5 +repage assets/img/host-$$h.png; done
	for h in $(hosts); do rm $$h-full.png; done
