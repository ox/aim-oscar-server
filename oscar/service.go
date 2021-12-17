package oscar

type Service interface {
	HandleSNAC(*Session, *SNAC) error
}
