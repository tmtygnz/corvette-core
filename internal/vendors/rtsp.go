package vendors

type RtspVendor struct {
	GenericVendor
}

func CreateRtspVendor(url string, cType string, name string) *RtspVendor {
	return &RtspVendor{
		GenericVendor: GenericVendor{
			camType: cType,
			url:     url,
			camName: name,
		},
	}
}

func (rv *RtspVendor) URL() string {
	return rv.url
}

func (rv *RtspVendor) Type() string {
	return rv.camType
}

func (rv *RtspVendor) CamName() string {
	return rv.camName
}
