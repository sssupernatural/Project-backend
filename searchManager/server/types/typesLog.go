package types

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetTypesLogger(out io.Writer)  {
	logger.Out = out
}