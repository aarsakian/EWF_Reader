package sections

type EWF_Done_Section struct {
}

func (done_section *EWF_Done_Section) Parse(buf []byte) {

	//defer Utils.TimeTrack(time.Now(), "Parsing Done Section")

}

func (done_section *EWF_Done_Section) GetAttr(string) interface{} {
	return ""
}
