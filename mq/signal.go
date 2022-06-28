package mq

// user id -> group topic -> complete
var CleanUpChans map[uint]map[string]chan bool

func init() {
	CleanUpChans = make(map[uint]map[string]chan bool)
}
