package keyword

type Adaptor struct{}

func (a *Adaptor) CheckContent(content string, channel_id int) (bool, string, error) {
	return false, "", nil
}
