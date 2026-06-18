package vendors

import "strconv"

type RtspVendor struct {
	GenericVendor
}

func CreateRtspVendor(id int, url string, surl string, cType string, name string) *RtspVendor {
	return &RtspVendor{
		GenericVendor: GenericVendor{
			id:      id,
			camType: cType,
			url:     url,
			surl:    surl,
			camName: name,
		},
	}
}

func (rv *RtspVendor) ID() int {
	return rv.id
}

func (rv *RtspVendor) IDStr() string {
	return strconv.Itoa(rv.id)
}

func (rv *RtspVendor) URL() string {
	return rv.url
}

func (rv *RtspVendor) SURL() string {
	return rv.surl
}

func (rv *RtspVendor) Type() string {
	return rv.camType
}

func (rv *RtspVendor) CamName() string {
	return rv.camName
}
