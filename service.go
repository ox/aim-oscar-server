package main

type Service interface {
	HandleSNAC(*Session, *SNAC) error
}
