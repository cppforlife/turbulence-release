package monit

import (
	"encoding/xml"
)

type status struct {
	XMLName  xml.Name `xml:"monit"`
	Services servicesTag
}

type servicesTag struct {
	XMLName  xml.Name     `xml:"services"`
	Services []serviceTag `xml:"service"`
}

type serviceTag struct {
	XMLName xml.Name `xml:"service"`
	Name    string   `xml:"name,attr"`
	PID     int      `xml:"pid"`
}
