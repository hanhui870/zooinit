package bootstrap

type WatchDog struct {
	env *envInfo

	internalClientUrl string

	discoveryClientUrl string
}

//
func NewWatchDog(env *envInfo, internal string, discovery string) *WatchDog {
	return &WatchDog{env: env, internalClientUrl: internal, discoveryClientUrl: discovery}
}

// TODO Watch dog run
func (w *WatchDog) Run() {

}
