package profile

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Source struct {
	Include []string
	Exclude []string
}

type Target struct {
	Filename      string
	PortraitSize  string `yaml:"portrait_size"`
	LandscapeSize string `yaml:"landscape_size"`
}

type Profile struct {
	Source Source
	Target Target
}

type outputContext struct {
	Stem string
	Ext  string
}

func (p *Profile) Matches(fname string) bool {
	matches := false
	for _, i := range p.Source.Include {
		r, err := regexp.Compile(i)
		if err != nil {
			return false
		}
		if !r.MatchString(fname) {
			return false
		} else {
			matches = true
			break
		}
	}

	for _, e := range p.Source.Exclude {
		r, err := regexp.Compile(e)
		if err != nil {
			return false
		}
		if r.MatchString(fname) {
			return false
		}
	}
	return matches
}

func (p *Profile) CalculateTargetFilename(inname string) (string, error) {
	ext := filepath.Ext(inname)
	c := outputContext{
		Ext:  ext,
		Stem: strings.TrimSuffix(inname, ext),
	}
	t, err := template.New("root").Parse(p.Target.Filename)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, c); err != nil {
		return "", err
	}
	return out.String(), nil
}
