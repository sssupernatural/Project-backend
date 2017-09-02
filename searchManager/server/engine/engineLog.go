package engine

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = logrus.New()

func SetEngineLogger(out io.Writer)  {
	logger.Out = out
}