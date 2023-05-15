package db

import "github.com/sirupsen/logrus"

var logger = logrus.StandardLogger().WithField("module", "db")
