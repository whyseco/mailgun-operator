package controller

import (
	"github.com/whyseco/mailgun-operator/pkg/controller/mailgunroute"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, mailgunroute.Add)
}
