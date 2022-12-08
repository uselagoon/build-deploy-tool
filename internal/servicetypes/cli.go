package servicetypes

var cli = ServiceType{
	Name: "cli",
}

var cliPersistent = ServiceType{
	Name:    "cli-persistent",
	Volumes: ServiceVolume{},
}
