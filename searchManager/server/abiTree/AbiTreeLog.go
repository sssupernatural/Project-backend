package abiTree

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetAbiTreeLogger(out io.Writer)  {
	logger.Out = out
}