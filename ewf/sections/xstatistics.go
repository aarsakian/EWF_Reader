package sections

type XStatistics struct {
	unknown []byte
}

func (xstats XStatistics) Parse(buf []byte) {
	xstats.unknown = buf
}

func (xstats XStatistics) GetAttr(attr string) interface{} {

	return ""
}
