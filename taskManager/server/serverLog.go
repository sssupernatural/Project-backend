package server

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetServerLogger(out io.Writer)  {
	logger.Out = out
}