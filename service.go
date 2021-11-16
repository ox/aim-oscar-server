package main

import (
	"context"
)

type Service interface {
	HandleSNAC(context.Context, *SNAC)
}
