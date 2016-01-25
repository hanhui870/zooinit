package utility

import "net"

type SRVService struct {
	// srv result
	srvs []*net.SRV
}

type SRVInfo struct {
	cname string
	ip    net.IP
	port  uint16
}

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
func NewSRVServiceOfDomainAndService(domain, service, protocal string) (*SRVService, error) {
	_, srvs, err := net.LookupSRV(service, protocal, domain)
	if err != nil {
		return nil, err
	}

	return &SRVService{srvs}, nil
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

	return nil, nil
}

func (s *SRVService) GetRandomService() (*SRVInfo, error) {

	return nil, nil
}
