package main

import "encoding/xml"

type NmaprunHeader struct {
	XMLName          xml.Name `xml:"nmaprun"`
	Text             string   `xml:",chardata"`
	Scanner          string   `xml:"scanner,attr"`
	Args             string   `xml:"args,attr"`
	Start            string   `xml:"start,attr"`
	Startstr         string   `xml:"startstr,attr"`
	Version          string   `xml:"version,attr"`
	Xmloutputversion string   `xml:"xmloutputversion,attr"`
}

type NmapPort struct {
	Text     string `xml:",chardata"`
	Protocol string `xml:"protocol,attr"`
	Portid   int    `xml:"portid,attr"`
	State    struct {
		Text      string `xml:",chardata"`
		State     string `xml:"state,attr"`
		Reason    string `xml:"reason,attr"`
		ReasonTtl string `xml:"reason_ttl,attr"`
	} `xml:"state"`
	Service struct {
		Text      string `xml:",chardata"`
		Name      string `xml:"name,attr"`
		Product   string `xml:"product,attr"`
		Version   string `xml:"version,attr"`
		Extrainfo string `xml:"extrainfo,attr"`
		Method    string `xml:"method,attr"`
		Conf      string `xml:"conf,attr"`
		Tunnel    string `xml:"tunnel,attr"`
		Hostname  string `xml:"hostname,attr"`
		Servicefp string `xml:"servicefp,attr"`
		Ostype    string `xml:"ostype,attr"`
	} `xml:"service"`
	Script []struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Output string `xml:"output,attr"`
	} `xml:"script"`
}

type NmapHost struct {
	Text   string `xml:",chardata"`
	Status struct {
		Text   string `xml:",chardata"`
		State  string `xml:"state,attr"`
		Reason string `xml:"reason,attr"`
	} `xml:"status"`
	Address struct {
		Text     string `xml:",chardata"`
		Addr     string `xml:"addr,attr"`
		Addrtype string `xml:"addrtype,attr"`
	} `xml:"address"`
	Hostnames struct {
		Text     string `xml:",chardata"`
		Hostname struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
			Type string `xml:"type,attr"`
		} `xml:"hostname"`
	} `xml:"hostnames"`
	Ports struct {
		Text       string `xml:",chardata"`
		Extraports []struct {
			Text         string `xml:",chardata"`
			State        string `xml:"state,attr"`
			Count        string `xml:"count,attr"`
			Extrareasons []struct {
				Text   string `xml:",chardata"`
				Reason string `xml:"reason,attr"`
				Count  string `xml:"count,attr"`
			} `xml:"extrareasons"`
		} `xml:"extraports"`
		Port []NmapPort `xml:"port"`
	} `xml:"ports"`
	Os struct {
		Text     string `xml:",chardata"`
		Portused []struct {
			Text   string `xml:",chardata"`
			State  string `xml:"state,attr"`
			Proto  string `xml:"proto,attr"`
			Portid string `xml:"portid,attr"`
		} `xml:"portused"`
		Osclass []struct {
			Text     string `xml:",chardata"`
			Type     string `xml:"type,attr"`
			Vendor   string `xml:"vendor,attr"`
			Osfamily string `xml:"osfamily,attr"`
			Osgen    string `xml:"osgen,attr"`
			Accuracy string `xml:"accuracy,attr"`
		} `xml:"osclass"`
		Osmatch []struct {
			Text     string `xml:",chardata"`
			Name     string `xml:"name,attr"`
			Accuracy string `xml:"accuracy,attr"`
			Line     string `xml:"line,attr"`
		} `xml:"osmatch"`
		Osfingerprint struct {
			Text        string `xml:",chardata"`
			Fingerprint string `xml:"fingerprint,attr"`
		} `xml:"osfingerprint"`
	} `xml:"os"`
	Uptime struct {
		Text     string `xml:",chardata"`
		Seconds  string `xml:"seconds,attr"`
		Lastboot string `xml:"lastboot,attr"`
	} `xml:"uptime"`
	Tcpsequence struct {
		Text       string `xml:",chardata"`
		Index      string `xml:"index,attr"`
		Class      string `xml:"class,attr"`
		Difficulty string `xml:"difficulty,attr"`
		Values     string `xml:"values,attr"`
	} `xml:"tcpsequence"`
	Ipidsequence struct {
		Text   string `xml:",chardata"`
		Class  string `xml:"class,attr"`
		Values string `xml:"values,attr"`
	} `xml:"ipidsequence"`
	Tcptssequence struct {
		Text   string `xml:",chardata"`
		Class  string `xml:"class,attr"`
		Values string `xml:"values,attr"`
	} `xml:"tcptssequence"`
	Trace struct {
		Text  string `xml:",chardata"`
		Port  string `xml:"port,attr"`
		Proto string `xml:"proto,attr"`
		Hop   []struct {
			Text   string `xml:",chardata"`
			Ttl    string `xml:"ttl,attr"`
			Rtt    string `xml:"rtt,attr"`
			Ipaddr string `xml:"ipaddr,attr"`
			Host   string `xml:"host,attr"`
		} `xml:"hop"`
	} `xml:"trace"`
	Times struct {
		Text   string `xml:",chardata"`
		Srtt   string `xml:"srtt,attr"`
		Rttvar string `xml:"rttvar,attr"`
		To     string `xml:"to,attr"`
	} `xml:"times"`
	Distance struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"distance"`
}

type Nmaprun struct {
	XMLName          xml.Name `xml:"nmaprun"`
	Text             string   `xml:",chardata"`
	Scanner          string   `xml:"scanner,attr"`
	Args             string   `xml:"args,attr"`
	Start            string   `xml:"start,attr"`
	Startstr         string   `xml:"startstr,attr"`
	Version          string   `xml:"version,attr"`
	Xmloutputversion string   `xml:"xmloutputversion,attr"`
	Scaninfo         struct {
		Text        string `xml:",chardata"`
		Type        string `xml:"type,attr"`
		Protocol    string `xml:"protocol,attr"`
		Numservices string `xml:"numservices,attr"`
		Services    string `xml:"services,attr"`
	} `xml:"scaninfo"`
	Verbose struct {
		Text  string `xml:",chardata"`
		Level string `xml:"level,attr"`
	} `xml:"verbose"`
	Debugging struct {
		Text  string `xml:",chardata"`
		Level string `xml:"level,attr"`
	} `xml:"debugging"`
	Taskbegin []struct {
		Text string `xml:",chardata"`
		Task string `xml:"task,attr"`
		Time string `xml:"time,attr"`
	} `xml:"taskbegin"`
	Taskend []struct {
		Text      string `xml:",chardata"`
		Task      string `xml:"task,attr"`
		Time      string `xml:"time,attr"`
		Extrainfo string `xml:"extrainfo,attr"`
	} `xml:"taskend"`
	Taskprogress []struct {
		Text      string `xml:",chardata"`
		Task      string `xml:"task,attr"`
		Time      string `xml:"time,attr"`
		Percent   string `xml:"percent,attr"`
		Remaining string `xml:"remaining,attr"`
		Etc       string `xml:"etc,attr"`
	} `xml:"taskprogress"`
	Hosts    []NmapHost `xml:"host"`
	Runstats struct {
		Text     string `xml:",chardata"`
		Finished struct {
			Text    string `xml:",chardata"`
			Time    string `xml:"time,attr"`
			Timestr string `xml:"timestr,attr"`
		} `xml:"finished"`
		Hosts struct {
			Text  string `xml:",chardata"`
			Up    string `xml:"up,attr"`
			Down  string `xml:"down,attr"`
			Total string `xml:"total,attr"`
		} `xml:"hosts"`
	} `xml:"runstats"`
}
