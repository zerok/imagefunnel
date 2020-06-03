package profile

import "regexp"

type Source struct {
	Include []string
	Exclude []string
}

type Profile struct {
	Source Source
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
