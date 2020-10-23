package cfg

import "github.com/gofoji/foji/stringlist"

func (c Config) Merge(from Config) Config {
	c.Formats = c.Formats.Merge(from.Formats)
	c.Processes = c.Processes.Merge(from.Processes)
	c.Processes = c.Processes.Formats(c.Formats)
	return c
}

func (pp Processes) Formats(formats Processes) Processes {
	if pp == nil {
		return pp
	}
	for key, p := range pp {
		f, ok := formats[p.Format]
		if ok {
			p = p.Merge(f)
		}
		pp[key] = p
	}
	return pp
}

func (pp Processes) Merge(from Processes) Processes {
	if pp == nil {
		pp = Processes{}
	} else {
		for key, p := range pp {
			p.ID = key
			pp[key] = p
		}

	}
	for key, p := range from {
		to, ok := pp[key]
		if ok {
			to = to.Merge(p)
		} else {
			to = p
		}
		to.ID = key
		pp[key] = to
	}
	return pp
}

func mergeOutputs(to, from stringlist.StringMap) stringlist.StringMap {
	if len(to) == 0 {
		return from
	}

	if len(to) == 1 {
		// Special case to disable inherited output
		if _, ok := to["-"]; ok {
			return stringlist.StringMap{}
		}
	}
	return to
}

func (o Output) Merge(from Output) Output {
	result := o
	if result == nil {
		result = Output{}
	}

	for k, v := range from {
		result[k] = mergeOutputs(result[k], v)
	}

	return result
}

func (p Process) Merge(from Process) Process {
	if p.Case == "" {
		p.Case = from.Case
	}

	if p.Format == "" {
		p.Format = from.Format
	}

	if len(p.Post) == 0 {
		p.Post = from.Post
	}

	if len(p.Resources) == 0 {
		p.Resources = from.Resources
	}

	p.Output = p.Output.Merge(from.Output)
	p.Maps = p.Maps.Merge(from.Maps)
	p.Params = p.Params.Merge(from.Params)

	return p
}

func (m Maps) Merge(from Maps) Maps {
	m.Type = MergeTypesMaps(from.Type, m.Type)
	m.Nullable = MergeTypesMaps(from.Nullable, m.Nullable)
	m.Name = MergeTypesMaps(from.Name, m.Name)
	m.Case = MergeTypesMaps(from.Case, m.Case)
	return m
}

func (pp ParamMap) Merge(from ParamMap) ParamMap {
	var out = pp
	if out == nil {
		out = ParamMap{}
	}

	for key, p := range from {
		_, ok := out[key]
		if !ok {
			out[key] = p
		}
	}
	return out
}
