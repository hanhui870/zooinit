package utility

import (
	"github.com/juju/errors"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

type SRVService struct {
	// srv result
	srvs []*net.SRV
}

type SRVInfo struct {
	ip   net.IP
	port uint16
}

type SRVList []*SRVInfo

// Fetch SRV dns records of A domain, use for discovery service
// in _etcd._tcp.discovery.alishui.com
func NewSRVServiceOfDomain(domain string) (*SRVService, error) {
	_, srvs, err := net.LookupSRV("", "", domain)
	if err != nil {
		return nil, err
	}

	return &SRVService{srvs}, nil
}

//Fetch SRV dns records of A domain, use for discovery service
// in discovery.alishui.com etcd tcp
func NewSRVServiceOfDomainAndService(service, protocal, domain string) (*SRVService, error) {
	_, srvs, err := net.LookupSRV(service, protocal, domain)
	if err != nil {
		return nil, err
	}

	return &SRVService{srvs}, nil
}

// Build New SRVInfo info
func NewSRVInfo(srv *net.SRV) (*SRVInfo, error) {
	sin := &SRVInfo{port: srv.Port}

	if ip := net.ParseIP(srv.Target[:len(srv.Target)-1]); ip != nil {
		sin.ip = ip
	} else if ips, err := net.LookupIP(srv.Target); err == nil {
		ip, err := GetRandomIpV4(ips)
		if err != nil {
			return nil, err
		}
		sin.ip = ip
	}

	return sin, nil
}

// Build New SRVInfo info
func NewSRVInfoBuild(ip string, port uint16) (*SRVInfo, error) {
	sin := &SRVInfo{port: port}

	if ipo := net.ParseIP(ip); ipo != nil {
		sin.ip = ipo
	} else {
		return nil, errors.New("IP parse for " + ip + " result nil.")
	}

	return sin, nil
}

// value.Weight , value.Priority Process
// Priority The priority of this target host. A client MUST attempt to contact the target host with the lowest-numbered priority it can reach;
// 	target hosts with the same priority SHOULD be tried in an order defined by the weight field. The range is 0-65535.
// 	This is a 16 bit unsigned integer in network byte order.
// Weight A server selection mechanism. The weight field specifies a relative weight for entries with the same priority.
// 	Larger weights SHOULD be given a proportionately higher probability of being selected. The range of this number is 0-65535.
// 	This is a 16 bit unsigned integer in network byte order. Domain administrators SHOULD use Weight 0 when there isn't any server selection to do,
// 	to make the RR easier to read for humans (less noisy). In the presence of records containing weights greater than 0,
// 	records with weight 0 should have a very small chance of being selected.
//
// Result of net.LookupSRV is sorted.
func (s *SRVService) GetRankedRandomService() (*SRVInfo, error) {
	if s == nil || s.srvs == nil {
		return nil, errors.New("SRVService object nil.")
	}

	return nil, nil
}

// return a random one
func (s *SRVService) GetRandomService() (*SRVInfo, error) {
	if s == nil || s.srvs == nil {
		return nil, errors.New("SRVService object nil.")
	}

	return NewSRVInfo(s.srvs[rand.Intn(len(s.srvs))])
}

// Fetch IPv4 result of dns result
// IPv4 default 16 length, with zero padding
func GetRandomIpV4(ips []net.IP) (net.IP, error) {
	var ipdst []net.IP
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			ipdst = append(ipdst, ip)
		}
	}

	if ipdst == nil {
		return nil, errors.New("Not found")
	}

	return ipdst[rand.Intn(len(ipdst))], nil
}

// Fetch IPv6 result of dns result
func GetRandomIpV6(ips []net.IP) (net.IP, error) {
	var ipdst []net.IP
	for _, ip := range ips {
		//ip.To16 no use. ip4==nil is IP6
		if ip4 := ip.To4(); ip4 == nil {
			ipdst = append(ipdst, ip)
		}
	}

	if ipdst == nil {
		return nil, errors.New("Not found")
	}

	return ipdst[rand.Intn(len(ipdst))], nil
}

// Format SRVList to ip1:port, ip2:port for cluster use
func (s SRVList) Endpoints() string {
	if s == nil {
		return ""
	}
	var ends []string

	for _, value := range s {
		tmp := value.ip.String() + ":" + strconv.Itoa(int(value.port))
		ends = append(ends, tmp)
	}

	return strings.Join(ends, ",")
}
