package core

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetCoreLogger(out io.Writer)  {
	logger.Out = out
}