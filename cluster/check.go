package cluster

// Health check
type ServiceCheck interface {
	IsHealth() bool

	Members()

	AddMember() error

	DelMember() error
}
