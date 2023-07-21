package servicetypes

var worker = ServiceType{
	Name: "worker",
}

var workerPersistent = ServiceType{
	Name:    "worker-persistent",
	Volumes: ServiceVolume{},
}
