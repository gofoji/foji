package cfg

import "github.com/gofoji/foji/stringlist"

// Merge merges all properties from an ancestor Config.
func (c Config) Merge(from Config) Config {
	c.Formats = c.Formats.Merge(from.Formats)
	c.Files = c.Files.Merge(from.Files)
	c.Processes = c.Processes.Merge(from.Processes).ApplyFormat(c.Formats)

	return c
}

// ApplyFormat merges linked format into each process config.
func (pp Processes) ApplyFormat(formats Processes) Processes {
	for key, p := range pp {
		f, ok := formats[p.Format]
		if ok {
			p = p.Merge(f)
		}

		pp[key] = p
	}

	return pp
}

// Merge merges all properties from an ancestor Processes.
func (pp Processes) Merge(from Processes) Processes {
	out := pp
	if out == nil {
		out = Processes{}
	} else {
		for key, p := range pp {
			p.ID = key
			out[key] = p
		}
	}

	for key, p := range from {
		to, ok := out[key]
		if ok {
			to = to.Merge(p)
		} else {
			to = p
		}

		to.ID = key
		out[key] = to
	}

	return out
}

// mergeOutputs merges all output lists, and has the special function to treat "-" as an empty array.
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

// Merge merges all properties from an ancestor Output.
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

// Merge merges all properties from an ancestor Process.
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

	p.Files = p.Files.Merge(from.Files)
	p.Output = p.Output.Merge(from.Output)
	p.Maps = p.Maps.Merge(from.Maps)
	p.Params = p.Params.Merge(from.Params)

	return p
}

// Merge merges all properties from an ancestor Maps.
func (m Maps) Merge(from Maps) Maps {
	m.Type = MergeTypesMaps(from.Type, m.Type)
	m.Nullable = MergeTypesMaps(from.Nullable, m.Nullable)
	m.Name = MergeTypesMaps(from.Name, m.Name)
	m.Case = MergeTypesMaps(from.Case, m.Case)

	return m
}

// Merge merges all properties from an ancestor ParamMap.
func (pp ParamMap) Merge(from ParamMap) ParamMap {
	out := pp
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

// Merge merges all properties from an ancestor FileInputMap.
func (ff FileInputMap) Merge(from FileInputMap) FileInputMap {
	out := ff
	if out == nil {
		out = FileInputMap{}
	}

	for key, p := range from {
		_, ok := out[key]
		if !ok {
			out[key] = p
		}
	}

	return out
}

// Merge merges all properties from an ancestor FileInput.
func (f FileInput) Merge(from FileInput) FileInput {
	if len(f.Files) > 0 || len(f.Filter) > 0 || len(f.Rewrite) > 0 {
		return f
	}

	return from
}

// MergeTypesMaps merges all properties from an ancestor TypeMap.
func MergeTypesMaps(maps ...stringlist.StringMap) stringlist.StringMap {
	result := stringlist.StringMap{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
