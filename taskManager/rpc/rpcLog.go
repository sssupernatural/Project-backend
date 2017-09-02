package rpc

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetRPCLogger(out io.Writer)  {
	logger.Out = out
}