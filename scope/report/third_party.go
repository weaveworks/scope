package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
)

// ThirdParty are rendered references to other systems, keyed on nodes
// (or origins) at any level of topology. All fields are configured by users at
// runtime, with the URL being rendered from a URL template.
type ThirdParty struct {
	Topology string `json:"topology"` // as named in app's topologyRegistry
	Label    string `json:"label"`    // what appears in the tab in the modal
	URL      string `json:"url"`      // what populates the iframe
	Iframe   bool   `json:"iframe"`   // if true, render as iframe, otherwise render as link
}

// RawThirdParty objects are the same as a ThirdParty, but URL is interpreted
// as a template. These are given in the config file.
type RawThirdParty ThirdParty

// ThirdPartyTemplate is a compiled ThirdParty object.
type ThirdPartyTemplate struct {
	Topology string
	Label    string
	Template *template.Template
	Iframe   bool
}

// ThirdPartyTemplates is a slice of ThirdPartyTemplates.
type ThirdPartyTemplates []ThirdPartyTemplate

// Compile compiles a ThirdParty object.
func (t ThirdParty) Compile(name string) (ThirdPartyTemplate, error) {
	tmpl, err := template.New(name).Parse(t.URL)
	if err != nil {
		return ThirdPartyTemplate{}, err
	}
	return ThirdPartyTemplate{
		Topology: t.Topology,
		Label:    t.Label,
		Template: tmpl,
		Iframe:   t.Iframe,
	}, nil
}

// ReadThirdPartyConf reads and compiles a file with ThirdParty definitions.
func ReadThirdPartyConf(file string) (ThirdPartyTemplates, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {

		return nil, err
	}
	var tpts []ThirdParty
	if err := json.Unmarshal(buf, &tpts); err != nil {
		return nil, err
	}
	var tps ThirdPartyTemplates
	for i, tp := range tpts {
		tmpl, err := tp.Compile(fmt.Sprintf("thirdparty_%d", i))
		if err != nil {
			return nil, err
		}
		tps = append(tps, tmpl)
	}
	return tps, nil
}

// For returns the ThirdParties relavant for the given topology name.
func (tps ThirdPartyTemplates) For(topology string) ThirdPartyTemplates {
	var res ThirdPartyTemplates
	for _, t := range tps {
		if t.Topology != topology {
			continue
		}
		res = append(res, t)
	}
	return res
}

// Execute renders the thirdparty templates for a given node.
func (tps ThirdPartyTemplates) Execute(node MappedNode) ([]ThirdParty, error) {
	var res []ThirdParty
	for _, t := range tps {
		var b = &bytes.Buffer{}
		if err := t.Template.Execute(b, node); err != nil {
			return nil, err
		}
		res = append(res, ThirdParty{
			Label:  t.Label,
			URL:    b.String(),
			Iframe: t.Iframe,
		})

	}
	return res, nil
}
